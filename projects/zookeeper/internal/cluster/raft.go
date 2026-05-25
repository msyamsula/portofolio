package cluster

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/syamsularifin/zookeeper/internal/wal"
)

// Storage is the interface RaftNode uses to access WAL and tree.
//
// Why an interface instead of a concrete Store?
//   - RaftNode lives in package "cluster", Store lives in package "store"
//   - An interface avoids tight coupling between the two
//   - Tests can use a simple in-memory implementation (no temp files)
//
// In production, *store.Store implements this interface.
// In tests, memoryStorage implements it.
type Storage interface {
	// AppendWAL writes an entry to the WAL (disk) and in-memory cache.
	// The entry must already have a TxID assigned.
	AppendWAL(entry wal.Entry) error

	// ApplyTree applies an entry to the in-memory tree.
	// Called only after the entry is committed (majority confirmed).
	ApplyTree(entry wal.Entry) error

	// GetWALEntriesFrom returns cached entries starting at fromTxID.
	// Used by the leader to grab entries for replication.
	GetWALEntriesFrom(fromTxID int64) []wal.Entry

	// LastWALTxID returns the TxID of the last WAL entry.
	// Returns 0 if no entries exist.
	LastWALTxID() int64

	// TruncateWALFrom removes all entries with TxID >= fromTxID.
	// Used when a follower has conflicting entries that don't match
	// the leader's log. The follower truncates and accepts the leader's version.
	TruncateWALFrom(fromTxID int64)
}

// RaftNode is the core Raft state machine.
//
// It knows:
//   - who it is and who the other nodes are (Config)
//   - what state it's in: follower, candidate, or leader (NodeState)
//
// It owns:
//   - a Storage (WAL + tree) for durability and state
//
// It can:
//   - handle AppendEntries messages (from leader)
//   - handle RequestVote messages (from candidates)
//   - propose new writes (leader only)
type RaftNode struct {
	mu     sync.Mutex
	config Config
	state  *NodeState
	logger *slog.Logger

	// store is the durable storage: WAL (disk) + tree (memory).
	// Leader writes to WAL during Propose.
	// Follower writes to WAL during HandleAppendEntries.
	// Both apply to tree after commit.
	store Storage

	// commitIndex is the highest TxID that's been committed
	// (replicated to a majority of nodes). Entries up to this
	// point are safe to apply to the state machine (tree).
	//
	// Only the leader advances this. Followers learn it from
	// the leader via AppendEntries.
	commitIndex int64

	// lastApplied is the highest TxID that's been applied to
	// the state machine (tree). Always <= commitIndex.
	//
	// The gap between lastApplied and commitIndex is "committed
	// but not yet applied." applyCommitted() closes this gap.
	lastApplied int64

	// transport is how we send messages to other nodes.
	transport Transport

	// heartbeatInterval is how often the leader sends heartbeats.
	// Default: 100ms. Short enough that followers don't time out.
	heartbeatInterval time.Duration

	// electionTimeout range. Each follower picks a random timeout
	// between min and max. The randomness prevents all followers
	// from starting elections at the same time (split vote).
	electionTimeoutMin time.Duration
	electionTimeoutMax time.Duration

	// lastHeartbeat is when we last heard from the leader.
	// If now - lastHeartbeat > electionTimeout → start election.
	lastHeartbeat time.Time

	// nextIndex tracks, for each peer, the next log entry the leader
	// will send to that peer. It's an optimistic guess — the leader
	// assumes everyone is caught up when it first wins election.
	// If a follower rejects, the leader jumps back using LastLogTxID.
	//
	// Only used when this node is the leader. nil otherwise.
	nextIndex map[NodeID]int64

	// matchIndex tracks, for each peer, the highest log entry that
	// the peer has confirmed it received. This is a fact, not a guess.
	// Only updated when a follower responds Success=true.
	//
	// Only used when this node is the leader. nil otherwise.
	matchIndex map[NodeID]int64

	// stopCh signals the loop to stop. Used for clean shutdown.
	stopCh chan struct{}
}

// NewRaftNode creates a node that starts as a follower in term 0.
func NewRaftNode(config Config, transport Transport, store Storage) *RaftNode {
	return &RaftNode{
		config:             config,
		state:              NewNodeState(),
		logger:             slog.New(slog.NewTextHandler(os.Stdout, nil)),
		transport:          transport,
		store:              store,
		heartbeatInterval:  100 * time.Millisecond,
		electionTimeoutMin: 300 * time.Millisecond,
		electionTimeoutMax: 500 * time.Millisecond,
		lastHeartbeat:      time.Now(),
		stopCh:             make(chan struct{}),
	}
}

// randomElectionTimeout returns a random duration between min and max.
// The randomness is critical: if all nodes used the same timeout,
// they'd all start elections at the same time → split vote → no leader.
//
// We use crypto/rand instead of math/rand because:
//   - math/rand is predictable (seeded by a deterministic source)
//   - crypto/rand uses the OS entropy pool (/dev/urandom)
//   - In a cluster, all nodes might start at the same time with similar
//     state. math/rand could produce similar sequences. crypto/rand won't.
func (rn *RaftNode) randomElectionTimeout() time.Duration {
	spread := int64(rn.electionTimeoutMax - rn.electionTimeoutMin)

	var b [8]byte
	rand.Read(b[:])
	n := int64(binary.LittleEndian.Uint64(b[:])) % spread
	if n < 0 {
		n = -n
	}

	return rn.electionTimeoutMin + time.Duration(n)
}

// Run starts the main loop in a goroutine.
// The loop does one thing: check what role I am and act accordingly.
//
//	Leader:    send heartbeats every 100ms
//	Follower:  if no heartbeat received for 300-500ms → start election
//	Candidate: same as follower (election timed out, try again)
func (rn *RaftNode) Run() {
	go rn.loop()
}

// Stop shuts down the loop.
func (rn *RaftNode) Stop() {
	close(rn.stopCh)
}

// loop is the main event loop. It runs forever until Stop is called.
//
// Every 50ms it wakes up and asks: "what should I do right now?"
//
//	If I'm the leader:
//	  Send heartbeats (with any new entries) to all followers
//
//	If I'm a follower or candidate:
//	  Has 300-500ms passed since I last heard from the leader?
//	  YES → start an election
func (rn *RaftNode) loop() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-rn.stopCh:
			return
		case <-ticker.C:
			rn.tick()
		}
	}
}

// tick is called every 50ms. It checks the current role and acts.
func (rn *RaftNode) tick() {
	rn.mu.Lock()
	role := rn.state.Role
	rn.mu.Unlock()

	switch role {
	case Leader:
		rn.leaderTick()
	case Follower, Candidate:
		rn.followerTick()
	}
}

// leaderTick sends entries (or heartbeats) to all followers.
//
// For each peer:
//  1. Look up nextIndex[peer] — "where is this peer in the log?"
//  2. Grab entries from the WAL starting at nextIndex
//  3. Send them via AppendEntries
//  4. On success → advance nextIndex and matchIndex
//  5. On failure → jump nextIndex to peer's LastLogTxID + 1
//
// If there are no new entries, this is just a heartbeat (empty Entries).
func (rn *RaftNode) leaderTick() {
	rn.mu.Lock()
	term := rn.state.CurrentTerm
	leaderID := rn.config.Self
	rn.mu.Unlock()

	for _, peer := range rn.config.OtherPeers() {
		rn.mu.Lock()
		next := rn.nextIndex[peer.ID]

		// Grab entries from WAL starting at nextIndex.
		// Returns nil if peer is caught up → heartbeat.
		entries := rn.store.GetWALEntriesFrom(next)

		// Compute prevLog: the entry right before the batch.
		// The follower checks this to verify log consistency.
		var prevLogTxID, prevLogTerm int64
		if next > 1 {
			prev := rn.store.GetWALEntriesFrom(next - 1)
			if prev != nil {
				prevLogTxID = prev[0].TxID
				prevLogTerm = prev[0].Term
			}
		}

		req := AppendEntriesRequest{
			Term:              term,
			LeaderID:          leaderID,
			PrevLogTxID:       prevLogTxID,
			PrevLogTerm:       prevLogTerm,
			Entries:           entries,
			LeaderCommitIndex: rn.commitIndex,
		}
		rn.mu.Unlock()

		resp, err := rn.transport.SendAppendEntries(peer, req)
		if err != nil {
			continue // peer unreachable, try next tick
		}

		rn.mu.Lock()
		// If peer has higher term, step down.
		if resp.Term > term {
			rn.becomeFollower(resp.Term, "")
			rn.mu.Unlock()
			return
		}

		if resp.Success && len(entries) > 0 {
			lastSent := entries[len(entries)-1].TxID
			rn.nextIndex[peer.ID] = lastSent + 1
			rn.matchIndex[peer.ID] = lastSent
		} else if !resp.Success {
			rn.nextIndex[peer.ID] = resp.LastLogTxID + 1
		}
		rn.mu.Unlock()
	}

	// After sending to all peers, check if we can advance commitIndex.
	rn.advanceCommitIndex()
}

// advanceCommitIndex checks if any new entries have been replicated
// to a majority. If so, advance commitIndex.
//
// The approach: collect all matchIndex values (including leader's own),
// sort descending, and pick the quorum-th value. That's the highest
// TxID that a majority of nodes have.
//
// Example (3-node cluster, quorum=2):
//
//	leader's LastWALTxID = 1000
//	matchIndex: node-1=700, node-3=500
//	All values: [1000, 700, 500]
//	Sort desc:  [1000, 700, 500]
//	Quorum-th (2nd): 700 → commitIndex = 700
//
// O(peers log peers) regardless of how many entries exist.
func (rn *RaftNode) advanceCommitIndex() {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	// Collect: leader's own position + all peers.
	matches := make([]int64, 0, len(rn.config.Peers))
	matches = append(matches, rn.store.LastWALTxID()) // leader has everything
	for _, peer := range rn.config.OtherPeers() {
		matches = append(matches, rn.matchIndex[peer.ID])
	}

	// Sort descending.
	sort.Slice(matches, func(i, j int) bool {
		return matches[i] > matches[j]
	})

	// The quorum-th value (0-indexed: QuorumSize()-1) is the highest TxID
	// that at least QuorumSize() nodes have.
	committed := matches[rn.config.QuorumSize()-1]

	if committed > rn.commitIndex {
		rn.commitIndex = committed
		rn.logger.Info("advanced commitIndex",
			"commitIndex", rn.commitIndex,
		)
	}

	rn.applyCommitted()
}

// applyCommitted applies all entries between lastApplied and commitIndex.
//
// This is where entries finally become real: CREATE actually creates a znode,
// SET actually updates data, DELETE actually removes a node.
//
// Called by both leader (after advanceCommitIndex) and follower
// (after learning commitIndex from the leader).
//
// Must be called with rn.mu held.
func (rn *RaftNode) applyCommitted() {
	if rn.lastApplied >= rn.commitIndex {
		return
	}

	// Grab all entries from lastApplied+1 onward, apply up to commitIndex.
	entries := rn.store.GetWALEntriesFrom(rn.lastApplied + 1)
	for _, entry := range entries {
		if entry.TxID > rn.commitIndex {
			break
		}
		rn.store.ApplyTree(entry)
		rn.lastApplied = entry.TxID
	}
}

// followerTick checks if we've timed out waiting for the leader.
func (rn *RaftNode) followerTick() {
	rn.mu.Lock()
	elapsed := time.Since(rn.lastHeartbeat)
	rn.mu.Unlock()

	timeout := rn.randomElectionTimeout()
	if elapsed < timeout {
		return // still within timeout, do nothing
	}

	// Timeout expired. Leader is probably dead. Start an election.
	rn.runElection()
}

// runElection runs a full election: start it, ask for votes, count them.
func (rn *RaftNode) runElection() {
	voteReq := rn.StartElection()
	votes := 1 // we already voted for ourselves in StartElection

	// Ask every other node for their vote
	for _, peer := range rn.config.OtherPeers() {
		resp, err := rn.transport.SendRequestVote(peer, voteReq)
		if err != nil {
			continue // peer unreachable, skip
		}

		won := rn.CollectVote(resp, &votes)
		if won {
			// We're leader now. Send immediate heartbeats so followers
			// know we exist and don't start their own elections.
			rn.leaderTick()
			return
		}
	}

	// Didn't get enough votes. Stay candidate.
	// Next tick will check timeout again and maybe retry.
}

// ResetElectionTimer is called when we receive a valid heartbeat
// from the leader. It pushes back the election timeout.
func (rn *RaftNode) ResetElectionTimer() {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.lastHeartbeat = time.Now()
}

// Propose accepts a new write from a client.
//
// Only the leader can accept writes. If this node isn't the leader,
// it returns an error — the client should retry on the leader.
//
// The flow:
//  1. Create entry in memory (NOT written to WAL yet)
//  2. Send to followers (they write to their WALs)
//  3. Count how many confirmed
//  4. Majority confirmed?
//     YES → write to leader's WAL + apply to tree → return success
//     NO  → discard entry → return error (nothing was written on leader)
//
// Why not write to WAL first?
// If we wrote to WAL and then consensus failed, the entry stays in
// the WAL. The tick loop would eventually commit it — but the client
// was told it failed. That's a lie. By writing AFTER consensus,
// error truly means "nothing happened."
func (rn *RaftNode) Propose(op wal.OpType, path string, data []byte) (wal.Entry, error) {
	rn.mu.Lock()
	if rn.state.Role != Leader {
		rn.mu.Unlock()
		return wal.Entry{}, fmt.Errorf("not the leader")
	}

	// Step 1: create entry in memory only. No WAL write yet.
	entry := wal.Entry{
		TxID: rn.store.LastWALTxID() + 1,
		Term: rn.state.CurrentTerm,
		Op:   op,
		Path: path,
		Data: data,
	}

	term := rn.state.CurrentTerm
	leaderID := rn.config.Self
	rn.mu.Unlock()

	rn.logger.Info("proposing entry",
		"txid", entry.TxID,
		"op", entry.Op,
		"path", entry.Path,
	)

	// Step 2: send to followers. Count leader as 1 success
	// (leader will write its WAL after consensus).
	successCount := 1

	for _, peer := range rn.config.OtherPeers() {
		rn.mu.Lock()
		next := rn.nextIndex[peer.ID]

		// Build entries to send: any catch-up entries + our new entry.
		var entries []wal.Entry
		if catchUp := rn.store.GetWALEntriesFrom(next); catchUp != nil {
			entries = append(entries, catchUp...)
		}
		entries = append(entries, entry)

		// Compute prevLog for consistency check.
		var prevLogTxID, prevLogTerm int64
		if next > 1 {
			prev := rn.store.GetWALEntriesFrom(next - 1)
			if prev != nil {
				prevLogTxID = prev[0].TxID
				prevLogTerm = prev[0].Term
			}
		}

		req := AppendEntriesRequest{
			Term:              term,
			LeaderID:          leaderID,
			PrevLogTxID:       prevLogTxID,
			PrevLogTerm:       prevLogTerm,
			Entries:           entries,
			LeaderCommitIndex: rn.commitIndex,
		}
		rn.mu.Unlock()

		resp, err := rn.transport.SendAppendEntries(peer, req)
		if err != nil {
			continue // peer unreachable
		}

		rn.mu.Lock()
		if resp.Term > term {
			rn.becomeFollower(resp.Term, "")
			rn.mu.Unlock()
			return wal.Entry{}, fmt.Errorf("lost leadership during replication")
		}

		if resp.Success {
			lastSent := entries[len(entries)-1].TxID
			rn.nextIndex[peer.ID] = lastSent + 1
			rn.matchIndex[peer.ID] = lastSent
			successCount++
		} else {
			rn.nextIndex[peer.ID] = resp.LastLogTxID + 1
		}
		rn.mu.Unlock()
	}

	// Step 3: check consensus.
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if successCount < rn.config.QuorumSize() {
		return wal.Entry{}, fmt.Errorf("no consensus: only %d/%d nodes confirmed",
			successCount, len(rn.config.Peers))
	}

	// Step 4: consensus achieved! Write to leader's WAL + commit + apply.
	if err := rn.store.AppendWAL(entry); err != nil {
		return wal.Entry{}, fmt.Errorf("WAL write failed: %w", err)
	}

	rn.commitIndex = entry.TxID
	rn.store.ApplyTree(entry)
	rn.lastApplied = entry.TxID

	rn.logger.Info("committed entry",
		"txid", entry.TxID,
		"commitIndex", rn.commitIndex,
	)

	return entry, nil
}

// appendEntry appends a new entry to the WAL without replicating.
// Used by tests that need manual control over replication timing.
func (rn *RaftNode) appendEntry(op wal.OpType, path string, data []byte) (wal.Entry, error) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if rn.state.Role != Leader {
		return wal.Entry{}, fmt.Errorf("not the leader")
	}

	entry := wal.Entry{
		TxID: rn.store.LastWALTxID() + 1,
		Term: rn.state.CurrentTerm,
		Op:   op,
		Path: path,
		Data: data,
	}

	if err := rn.store.AppendWAL(entry); err != nil {
		return wal.Entry{}, fmt.Errorf("WAL write failed: %w", err)
	}

	rn.logger.Info("proposed entry",
		"txid", entry.TxID,
		"op", entry.Op,
		"path", entry.Path,
	)

	return entry, nil
}

// GetState returns a copy of the current state. Safe for reading from outside.
func (rn *RaftNode) GetState() NodeState {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return *rn.state
}

// GetCommitIndex returns the current commitIndex.
func (rn *RaftNode) GetCommitIndex() int64 {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.commitIndex
}

// HandleAppendEntries processes an AppendEntries message from a leader.
//
// This is the most common message in Raft. The leader sends it:
//   - Every ~150ms as a heartbeat (empty entries)
//   - Whenever there are new WAL entries to replicate
//
// The logic:
//
//  1. Is the sender's term < my term?
//     YES → reject. The sender is a stale leader.
//
//  2. Is the sender's term >= my term?
//     YES → accept. This is a legitimate leader.
//     → Update my term to match
//     → Step down to follower (if I was candidate or leader)
//     → Record who the leader is
//
//  3. Write new entries to our WAL (skip entries we already have).
//  4. Update commitIndex from the leader.
//  5. Apply any newly committed entries to the tree.
func (rn *RaftNode) HandleAppendEntries(req AppendEntriesRequest) AppendEntriesResponse {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	// Rule 1: reject if the sender's term is old.
	if req.Term < rn.state.CurrentTerm {
		rn.logger.Info("rejecting AppendEntries: stale term",
			"from", req.LeaderID,
			"their_term", req.Term,
			"my_term", rn.state.CurrentTerm,
		)
		return AppendEntriesResponse{
			Term:        rn.state.CurrentTerm,
			Success:     false,
			LastLogTxID: rn.store.LastWALTxID(),
		}
	}

	// Rule 2: accept. The sender's term is >= ours.
	rn.becomeFollower(req.Term, req.LeaderID)

	// Reset the election timer.
	rn.lastHeartbeat = time.Now()

	// Rule 3: check log consistency using prevLog.
	//
	// The leader says: "the entry before my batch should be TxID=X, Term=Y."
	// We check ONE entry. If it matches, our logs agree up to that point.
	// If not, reject — the leader will back up and retry.
	//
	// This is O(1) — no scanning of all entries.
	if req.PrevLogTxID > 0 {
		prevEntries := rn.store.GetWALEntriesFrom(req.PrevLogTxID)
		if prevEntries == nil || prevEntries[0].TxID != req.PrevLogTxID {
			// We don't have the previous entry. Our log is shorter.
			rn.logger.Info("rejecting AppendEntries: missing prevLog",
				"prevLogTxID", req.PrevLogTxID,
				"myLastTxID", rn.store.LastWALTxID(),
			)
			return AppendEntriesResponse{
				Term:        rn.state.CurrentTerm,
				Success:     false,
				LastLogTxID: rn.store.LastWALTxID(),
			}
		}
		if prevEntries[0].Term != req.PrevLogTerm {
			// We have the entry but from a different term — conflict.
			// Truncate from here so the leader can resend the correct entries.
			rn.logger.Info("rejecting AppendEntries: prevLog term mismatch",
				"prevLogTxID", req.PrevLogTxID,
				"expected_term", req.PrevLogTerm,
				"actual_term", prevEntries[0].Term,
			)
			rn.store.TruncateWALFrom(req.PrevLogTxID)
			return AppendEntriesResponse{
				Term:        rn.state.CurrentTerm,
				Success:     false,
				LastLogTxID: rn.store.LastWALTxID(),
			}
		}
	}

	// Rule 4: prevLog matches (or PrevLogTxID=0, meaning "from the start").
	// Truncate any stale entries after prevLog, then append the leader's entries.
	if len(req.Entries) > 0 {
		rn.store.TruncateWALFrom(req.Entries[0].TxID)
		for _, entry := range req.Entries {
			rn.store.AppendWAL(entry)
		}
	}

	// Rule 5: update commitIndex from the leader.
	// Use min(LeaderCommitIndex, lastWALTxID) because we can't commit
	// entries we don't have yet.
	lastTxID := rn.store.LastWALTxID()
	if req.LeaderCommitIndex > rn.commitIndex {
		rn.commitIndex = req.LeaderCommitIndex
		if rn.commitIndex > lastTxID {
			rn.commitIndex = lastTxID
		}
	}

	// Rule 6: apply any newly committed entries.
	rn.applyCommitted()

	return AppendEntriesResponse{
		Term:        rn.state.CurrentTerm,
		Success:     true,
		LastLogTxID: rn.store.LastWALTxID(),
	}
}

// HandleRequestVote processes a vote request from a candidate.
//
// A candidate sends this to all nodes when it starts an election.
// The voter decides:
//
//  1. Is the candidate's term < my term?
//     YES → reject. The candidate is behind.
//
//  2. Have I already voted for someone else this term?
//     YES → reject. One vote per term.
//
//  3. Is the candidate's log less up-to-date than mine?
//     YES → reject. Electing it could lose committed data.
//
//  4. Otherwise → grant the vote.
func (rn *RaftNode) HandleRequestVote(req RequestVoteRequest) RequestVoteResponse {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	// Rule 1: reject if the candidate's term is old.
	if req.Term < rn.state.CurrentTerm {
		rn.logger.Info("rejecting vote: stale term",
			"from", req.CandidateID,
			"their_term", req.Term,
			"my_term", rn.state.CurrentTerm,
		)
		return RequestVoteResponse{
			Term:        rn.state.CurrentTerm,
			VoteGranted: false,
		}
	}

	// If the candidate's term is higher than ours, update our term
	// and clear our vote (new term = new election = can vote again).
	if req.Term > rn.state.CurrentTerm {
		rn.becomeFollower(req.Term, "")
	}

	// Rule 2: have I already voted for someone else this term?
	alreadyVoted := rn.state.VotedFor != "" && rn.state.VotedFor != req.CandidateID
	if alreadyVoted {
		rn.logger.Info("rejecting vote: already voted",
			"from", req.CandidateID,
			"voted_for", rn.state.VotedFor,
		)
		return RequestVoteResponse{
			Term:        rn.state.CurrentTerm,
			VoteGranted: false,
		}
	}

	// Rule 3: is the candidate's log at least as up-to-date as mine?
	// A candidate with fewer entries might be missing committed data.
	if req.LastLogTxID < rn.store.LastWALTxID() {
		rn.logger.Info("rejecting vote: candidate log behind",
			"from", req.CandidateID,
			"their_last_tx", req.LastLogTxID,
			"my_last_tx", rn.store.LastWALTxID(),
		)
		return RequestVoteResponse{
			Term:        rn.state.CurrentTerm,
			VoteGranted: false,
		}
	}

	// Rule 4: grant the vote.
	rn.state.VotedFor = req.CandidateID
	rn.logger.Info("granting vote",
		"to", req.CandidateID,
		"term", rn.state.CurrentTerm,
	)

	return RequestVoteResponse{
		Term:        rn.state.CurrentTerm,
		VoteGranted: true,
	}
}

// StartElection is called when a follower hasn't heard from the leader
// for too long. It transitions to candidate and prepares a vote request.
//
// What happens step by step:
//
//  1. Increment term (new election round)
//  2. Become candidate
//  3. Vote for myself
//  4. Return the vote request to send to other nodes
//
// This method does NOT send the request — it just prepares it.
// The caller (network layer) is responsible for actually sending it
// to other nodes and collecting responses.
func (rn *RaftNode) StartElection() RequestVoteRequest {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	// Step 1: new term
	rn.state.CurrentTerm++

	// Step 2: become candidate
	rn.state.Role = Candidate
	rn.state.LeaderID = "" // no leader during election

	// Step 3: vote for myself
	rn.state.VotedFor = rn.config.Self

	// Step 3b: reset election timer so we don't immediately start
	// another election on the next tick. We wait a full random timeout
	// before trying again if this election fails.
	rn.lastHeartbeat = time.Now()

	rn.logger.Info("starting election",
		"node", rn.config.Self,
		"term", rn.state.CurrentTerm,
	)

	// Step 4: build the request for others
	return RequestVoteRequest{
		Term:        rn.state.CurrentTerm,
		CandidateID: rn.config.Self,
		LastLogTxID: rn.store.LastWALTxID(),
	}
}

// CollectVote processes one vote response and returns true if we've won.
//
// The candidate calls this for each response it receives.
// It tracks how many votes it has. When it reaches quorum, it wins.
//
// Returns:
//
//	won=true  → we got enough votes, we're now leader
//	won=false → not enough votes yet (or we lost)
func (rn *RaftNode) CollectVote(resp RequestVoteResponse, votes *int) (won bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	// If we're no longer a candidate, ignore the vote.
	// This can happen if we received a higher-term message
	// while waiting for votes and stepped down to follower.
	if rn.state.Role != Candidate {
		return false
	}

	// If the voter's term is higher, step down.
	// Someone else started a newer election.
	if resp.Term > rn.state.CurrentTerm {
		rn.becomeFollower(resp.Term, "")
		return false
	}

	// Count the vote
	if resp.VoteGranted {
		*votes++
	}

	// Check if we have enough votes
	if *votes >= rn.config.QuorumSize() {
		rn.becomeLeader()
		return true
	}

	return false
}

// becomeLeader transitions to leader state.
//
// Initializes nextIndex and matchIndex for all peers.
//   - nextIndex: start at LastWALTxID + 1 (optimistic: assume everyone is caught up)
//   - matchIndex: start at 0 (pessimistic: we haven't confirmed anything yet)
//
// If the guess is wrong (follower is behind), the follower will reject
// and the leader will jump nextIndex using LastLogTxID. This converges quickly.
func (rn *RaftNode) becomeLeader() {
	rn.state.Role = Leader
	rn.state.LeaderID = rn.config.Self

	lastTxID := rn.store.LastWALTxID()
	rn.nextIndex = make(map[NodeID]int64)
	rn.matchIndex = make(map[NodeID]int64)
	for _, peer := range rn.config.OtherPeers() {
		rn.nextIndex[peer.ID] = lastTxID + 1
		rn.matchIndex[peer.ID] = 0
	}

	rn.logger.Info("became leader",
		"node", rn.config.Self,
		"term", rn.state.CurrentTerm,
	)
}

// becomeFollower transitions to follower state.
// Called when we see a higher term or receive a valid AppendEntries.
func (rn *RaftNode) becomeFollower(term int64, leaderID NodeID) {
	oldRole := rn.state.Role
	oldTerm := rn.state.CurrentTerm

	rn.state.Role = Follower
	rn.state.LeaderID = leaderID

	// Clear vote if we moved to a new term.
	// New term = new election = can vote again.
	// Must compare BEFORE updating CurrentTerm.
	if term > oldTerm {
		rn.state.VotedFor = ""
	}

	rn.state.CurrentTerm = term

	if oldRole != Follower {
		rn.logger.Info("became follower",
			"term", term,
			"leader", leaderID,
			"was", fmt.Sprintf("%s", oldRole),
		)
	}
}
