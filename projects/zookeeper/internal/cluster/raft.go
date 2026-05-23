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

// RaftNode is the core Raft state machine.
//
// It knows:
//   - who it is and who the other nodes are (Config)
//   - what state it's in: follower, candidate, or leader (NodeState)
//
// It can:
//   - handle AppendEntries messages (from leader)
//   - handle RequestVote messages (from candidates)
//
// No networking here. No timers. Just: "given this message, what should I do?"
// We add networking and timers in later steps.
type RaftNode struct {
	mu     sync.Mutex
	config Config
	state  *NodeState
	logger *slog.Logger

	// log is the in-memory list of WAL entries this node knows about.
	// The leader appends new entries here via Propose().
	// Followers will receive entries via AppendEntries (later step).
	log []wal.Entry

	// lastLogTxID tracks the TxID of our latest log entry.
	// Used in elections: voters won't vote for a candidate with
	// a shorter log (fewer entries) than their own.
	lastLogTxID int64

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

	// applyFunc is called for each committed entry to apply it
	// to the state machine. The Store will set this to a function
	// that executes CREATE/SET/DELETE on the tree.
	//
	// If nil, entries are committed but not applied (useful for tests
	// that only care about replication, not application).
	applyFunc func(entry wal.Entry)

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
	// If a follower rejects, the leader decrements and retries.
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
func NewRaftNode(config Config, transport Transport) *RaftNode {
	return &RaftNode{
		config:             config,
		state:              NewNodeState(),
		logger:             slog.New(slog.NewTextHandler(os.Stdout, nil)),
		transport:          transport,
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
//   Leader:    send heartbeats every 100ms
//   Follower:  if no heartbeat received for 300-500ms → start election
//   Candidate: same as follower (election timed out, try again)
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
//   If I'm the leader:
//     Has 100ms passed since my last heartbeat?
//     YES → send heartbeats to all followers
//
//   If I'm a follower or candidate:
//     Has 300-500ms passed since I last heard from the leader?
//     YES → start an election
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
//   1. Look up nextIndex[peer] — "where is this peer in the log?"
//   2. Grab all entries from nextIndex onward
//   3. Send them via AppendEntries
//   4. On success → advance nextIndex and matchIndex
//   5. On failure → decrement nextIndex (peer is further behind than we thought)
//
// If there are no new entries, this is just a heartbeat (empty Entries).
func (rn *RaftNode) leaderTick() {
	rn.mu.Lock()
	term := rn.state.CurrentTerm
	leaderID := rn.config.Self
	rn.mu.Unlock()

	for _, peer := range rn.config.OtherPeers() {
		// Step 1: figure out what entries this peer needs.
		rn.mu.Lock()
		next := rn.nextIndex[peer.ID]

		// Step 2: grab entries from nextIndex onward.
		// next is a TxID (1-based). log is a slice (0-based).
		// Entry with TxID=1 is at log[0], TxID=2 at log[1], etc.
		// So entries to send start at log[next-1].
		var entries []wal.Entry
		if idx := int(next - 1); idx < len(rn.log) {
			entries = rn.log[idx:]
		}
		// If idx >= len(log), entries stays nil → heartbeat.

		req := AppendEntriesRequest{
			Term:              term,
			LeaderID:          leaderID,
			Entries:           entries,
			LeaderCommitIndex: rn.commitIndex,
		}
		rn.mu.Unlock()

		// Step 3: send to peer.
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
			// Step 4: peer accepted. Update tracking.
			lastSent := entries[len(entries)-1].TxID
			rn.nextIndex[peer.ID] = lastSent + 1
			rn.matchIndex[peer.ID] = lastSent
		} else if !resp.Success {
			// Step 5: peer rejected. Jump to where the peer actually is.
			// Old way:  nextIndex-- (one at a time, slow)
			// New way:  nextIndex = peer's LastLogTxID + 1 (one round trip)
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
//   leader's lastLogTxID = 1000
//   matchIndex: node-1=700, node-3=500
//   All values: [1000, 700, 500]
//   Sort desc:  [1000, 700, 500]
//   Quorum-th (2nd): 700 → commitIndex = 700
//
// O(peers log peers) regardless of how many entries exist.
func (rn *RaftNode) advanceCommitIndex() {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	// Collect: leader's own position + all peers.
	matches := make([]int64, 0, len(rn.config.Peers))
	matches = append(matches, rn.lastLogTxID) // leader has everything
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
	if rn.applyFunc == nil {
		return
	}

	for rn.lastApplied < rn.commitIndex {
		rn.lastApplied++
		// log is 0-indexed, TxID is 1-indexed.
		entry := rn.log[rn.lastApplied-1]
		rn.applyFunc(entry)
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

// SetApplyFunc sets the function that's called when an entry is committed.
// The Store uses this to execute CREATE/SET/DELETE on the tree.
func (rn *RaftNode) SetApplyFunc(fn func(entry wal.Entry)) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.applyFunc = fn
}

// Propose accepts a new write from a client.
//
// Only the leader can accept writes. If this node isn't the leader,
// it returns an error — the client should retry on the leader.
//
// What happens:
//   1. Assign a TxID (lastLogTxID + 1)
//   2. Append the entry to our local log
//   3. Return the entry (caller will replicate it later)
//
// Note: the entry is NOT committed yet. It's just in the leader's log.
// It becomes committed only after a majority of nodes have it (later step).
func (rn *RaftNode) Propose(op wal.OpType, path string, data []byte) (wal.Entry, error) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if rn.state.Role != Leader {
		return wal.Entry{}, fmt.Errorf("not the leader")
	}

	rn.lastLogTxID++
	entry := wal.Entry{
		TxID: rn.lastLogTxID,
		Op:   op,
		Path: path,
		Data: data,
	}
	rn.log = append(rn.log, entry)

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
//   1. Is the sender's term < my term?
//      YES → reject. The sender is a stale leader.
//
//   2. Is the sender's term >= my term?
//      YES → accept. This is a legitimate leader.
//      → Update my term to match
//      → Step down to follower (if I was candidate or leader)
//      → Record who the leader is
//
// That's the heartbeat logic. Now also handles log replication:
//   3. Append any new entries from the leader to our log.
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
			LastLogTxID: rn.lastLogTxID,
		}
	}

	// Rule 2: accept. The sender's term is >= ours.
	rn.becomeFollower(req.Term, req.LeaderID)

	// Reset the election timer.
	rn.lastHeartbeat = time.Now()

	// Rule 3: append new entries to our log.
	// The leader already knows our lastLogTxID (from our response),
	// so it only sends entries we don't have. Just append all of them.
	if len(req.Entries) > 0 {
		rn.log = append(rn.log, req.Entries...)
		rn.lastLogTxID = req.Entries[len(req.Entries)-1].TxID
	}

	// Rule 4: update commitIndex from the leader.
	// Use min(LeaderCommitIndex, lastLogTxID) because we can't commit
	// entries we don't have yet.
	if req.LeaderCommitIndex > rn.commitIndex {
		rn.commitIndex = req.LeaderCommitIndex
		if rn.commitIndex > rn.lastLogTxID {
			rn.commitIndex = rn.lastLogTxID
		}
	}

	// Rule 5: apply any newly committed entries.
	rn.applyCommitted()

	return AppendEntriesResponse{
		Term:        rn.state.CurrentTerm,
		Success:     true,
		LastLogTxID: rn.lastLogTxID,
	}
}

// HandleRequestVote processes a vote request from a candidate.
//
// A candidate sends this to all nodes when it starts an election.
// The voter decides:
//
//   1. Is the candidate's term < my term?
//      YES → reject. The candidate is behind.
//
//   2. Have I already voted for someone else this term?
//      YES → reject. One vote per term.
//
//   3. Is the candidate's log less up-to-date than mine?
//      YES → reject. Electing it could lose committed data.
//
//   4. Otherwise → grant the vote.
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
	if req.LastLogTxID < rn.lastLogTxID {
		rn.logger.Info("rejecting vote: candidate log behind",
			"from", req.CandidateID,
			"their_last_tx", req.LastLogTxID,
			"my_last_tx", rn.lastLogTxID,
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
//   1. Increment term (new election round)
//   2. Become candidate
//   3. Vote for myself
//   4. Return the vote request to send to other nodes
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
		LastLogTxID: rn.lastLogTxID,
	}
}

// CollectVote processes one vote response and returns true if we've won.
//
// The candidate calls this for each response it receives.
// It tracks how many votes it has. When it reaches quorum, it wins.
//
// Returns:
//   won=true  → we got enough votes, we're now leader
//   won=false → not enough votes yet (or we lost)
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
//   - nextIndex: start at lastLogTxID + 1 (optimistic: assume everyone is caught up)
//   - matchIndex: start at 0 (pessimistic: we haven't confirmed anything yet)
//
// If the guess is wrong (follower is behind), the follower will reject
// and the leader will decrement nextIndex and retry. This converges quickly.
func (rn *RaftNode) becomeLeader() {
	rn.state.Role = Leader
	rn.state.LeaderID = rn.config.Self

	// Initialize replication tracking for each peer.
	rn.nextIndex = make(map[NodeID]int64)
	rn.matchIndex = make(map[NodeID]int64)
	for _, peer := range rn.config.OtherPeers() {
		rn.nextIndex[peer.ID] = rn.lastLogTxID + 1
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
