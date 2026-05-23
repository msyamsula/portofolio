package cluster

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
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

	// lastLogTxID tracks the TxID of our latest WAL entry.
	// Used in elections: voters won't vote for a candidate with
	// a shorter log (fewer entries) than their own.
	lastLogTxID int64
}

// NewRaftNode creates a node that starts as a follower in term 0.
func NewRaftNode(config Config) *RaftNode {
	return &RaftNode{
		config: config,
		state:  NewNodeState(),
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

// GetState returns a copy of the current state. Safe for reading from outside.
func (rn *RaftNode) GetState() NodeState {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return *rn.state
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
// That's the entire heartbeat logic. Log replication (processing entries)
// comes in a later step.
func (rn *RaftNode) HandleAppendEntries(req AppendEntriesRequest) AppendEntriesResponse {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	// Rule 1: reject if the sender's term is old.
	// This means the sender is a stale leader — it was leader in a previous
	// term but doesn't know it's been replaced.
	if req.Term < rn.state.CurrentTerm {
		rn.logger.Info("rejecting AppendEntries: stale term",
			"from", req.LeaderID,
			"their_term", req.Term,
			"my_term", rn.state.CurrentTerm,
		)
		return AppendEntriesResponse{
			Term:    rn.state.CurrentTerm,
			Success: false,
		}
	}

	// Rule 2: accept. The sender's term is >= ours.
	// This means the sender is a legitimate leader (or at least newer than us).
	rn.becomeFollower(req.Term, req.LeaderID)

	return AppendEntriesResponse{
		Term:    rn.state.CurrentTerm,
		Success: true,
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
func (rn *RaftNode) becomeLeader() {
	rn.state.Role = Leader
	rn.state.LeaderID = rn.config.Self

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
