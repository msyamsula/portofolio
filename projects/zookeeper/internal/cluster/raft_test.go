package cluster

import "testing"

func newTestNode(id NodeID) *RaftNode {
	return NewRaftNode(Config{
		Self: id,
		Peers: []Peer{
			{ID: "node-1", Addr: "localhost:3001"},
			{ID: "node-2", Addr: "localhost:3002"},
			{ID: "node-3", Addr: "localhost:3003"},
		},
	})
}

// --- AppendEntries tests ---

func TestAppendEntries_AcceptFromLeader(t *testing.T) {
	node := newTestNode("node-1")

	// Leader in term 1 sends heartbeat
	resp := node.HandleAppendEntries(AppendEntriesRequest{
		Term:     1,
		LeaderID: "node-2",
	})

	if !resp.Success {
		t.Fatal("should accept AppendEntries from valid leader")
	}

	state := node.GetState()
	if state.CurrentTerm != 1 {
		t.Fatalf("expected term 1, got %d", state.CurrentTerm)
	}
	if state.LeaderID != "node-2" {
		t.Fatalf("expected leader node-2, got %s", state.LeaderID)
	}
}

func TestAppendEntries_RejectStaleTerm(t *testing.T) {
	node := newTestNode("node-1")

	// First, node learns about term 5
	node.HandleAppendEntries(AppendEntriesRequest{Term: 5, LeaderID: "node-2"})

	// Then a stale leader from term 3 sends a heartbeat
	resp := node.HandleAppendEntries(AppendEntriesRequest{Term: 3, LeaderID: "node-3"})

	if resp.Success {
		t.Fatal("should reject AppendEntries from stale term")
	}
	if resp.Term != 5 {
		t.Fatalf("response should carry current term 5, got %d", resp.Term)
	}
}

// --- RequestVote tests ---

func TestRequestVote_GrantVote(t *testing.T) {
	node := newTestNode("node-1")

	resp := node.HandleRequestVote(RequestVoteRequest{
		Term:        1,
		CandidateID: "node-2",
		LastLogTxID: 0,
	})

	if !resp.VoteGranted {
		t.Fatal("should grant vote to valid candidate")
	}

	state := node.GetState()
	if state.VotedFor != "node-2" {
		t.Fatalf("expected VotedFor=node-2, got %s", state.VotedFor)
	}
}

func TestRequestVote_RejectStaleTerm(t *testing.T) {
	node := newTestNode("node-1")

	// Node is in term 5
	node.HandleAppendEntries(AppendEntriesRequest{Term: 5, LeaderID: "node-3"})

	// Candidate from term 3 asks for vote
	resp := node.HandleRequestVote(RequestVoteRequest{
		Term:        3,
		CandidateID: "node-2",
		LastLogTxID: 0,
	})

	if resp.VoteGranted {
		t.Fatal("should reject vote from stale term")
	}
}

func TestRequestVote_RejectAlreadyVoted(t *testing.T) {
	node := newTestNode("node-1")

	// Vote for node-2 in term 1
	node.HandleRequestVote(RequestVoteRequest{
		Term:        1,
		CandidateID: "node-2",
		LastLogTxID: 0,
	})

	// node-3 also asks for vote in term 1
	resp := node.HandleRequestVote(RequestVoteRequest{
		Term:        1,
		CandidateID: "node-3",
		LastLogTxID: 0,
	})

	if resp.VoteGranted {
		t.Fatal("should reject: already voted for node-2 this term")
	}
}

func TestRequestVote_NewTermClearsVote(t *testing.T) {
	node := newTestNode("node-1")

	// Vote for node-2 in term 1
	node.HandleRequestVote(RequestVoteRequest{
		Term:        1,
		CandidateID: "node-2",
		LastLogTxID: 0,
	})

	// node-3 asks for vote in term 2 — new term, can vote again
	resp := node.HandleRequestVote(RequestVoteRequest{
		Term:        2,
		CandidateID: "node-3",
		LastLogTxID: 0,
	})

	if !resp.VoteGranted {
		t.Fatal("should grant vote in new term")
	}
}

func TestRequestVote_RejectCandidateWithShorterLog(t *testing.T) {
	node := newTestNode("node-1")
	// Simulate this node having entries up to TxID 10
	node.mu.Lock()
	node.lastLogTxID = 10
	node.mu.Unlock()

	// Candidate only has entries up to TxID 5
	resp := node.HandleRequestVote(RequestVoteRequest{
		Term:        1,
		CandidateID: "node-2",
		LastLogTxID: 5,
	})

	if resp.VoteGranted {
		t.Fatal("should reject candidate with shorter log")
	}
}
