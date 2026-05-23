package cluster

import (
	"fmt"
	"testing"
	"time"
)

// fakeTransport connects nodes directly via method calls. No network needed.
// It holds a map of all nodes so it can forward messages to the right one.
type fakeTransport struct {
	nodes map[NodeID]*RaftNode
}

func (ft *fakeTransport) SendRequestVote(peer Peer, req RequestVoteRequest) (RequestVoteResponse, error) {
	node, ok := ft.nodes[peer.ID]
	if !ok {
		return RequestVoteResponse{}, fmt.Errorf("node %s not found", peer.ID)
	}
	return node.HandleRequestVote(req), nil
}

func (ft *fakeTransport) SendAppendEntries(peer Peer, req AppendEntriesRequest) (AppendEntriesResponse, error) {
	node, ok := ft.nodes[peer.ID]
	if !ok {
		return AppendEntriesResponse{}, fmt.Errorf("node %s not found", peer.ID)
	}
	return node.HandleAppendEntries(req), nil
}

var testPeers = []Peer{
	{ID: "node-1", Addr: "localhost:3001"},
	{ID: "node-2", Addr: "localhost:3002"},
	{ID: "node-3", Addr: "localhost:3003"},
}

func newTestNode(id NodeID) *RaftNode {
	// Use a no-op transport for unit tests that don't need networking
	return NewRaftNode(Config{Self: id, Peers: testPeers}, &fakeTransport{})
}

// newTestCluster creates 3 connected nodes that can talk to each other.
func newTestCluster() (map[NodeID]*RaftNode, *fakeTransport) {
	ft := &fakeTransport{nodes: make(map[NodeID]*RaftNode)}

	for _, p := range testPeers {
		node := NewRaftNode(Config{Self: p.ID, Peers: testPeers}, ft)
		ft.nodes[p.ID] = node
	}

	return ft.nodes, ft
}

// ignore the unused warning for time in tests that don't use it
var _ = time.Now

// --- Propose tests ---

func TestPropose_LeaderAccepts(t *testing.T) {
	nodes, _ := newTestCluster()
	node2 := nodes["node-2"]

	// Make node-2 the leader via election
	voteReq := node2.StartElection()
	for _, peer := range node2.config.OtherPeers() {
		resp := nodes[peer.ID].HandleRequestVote(voteReq)
		votes := 1
		node2.CollectVote(resp, &votes)
	}

	state := node2.GetState()
	if state.Role != Leader {
		t.Fatal("node-2 should be leader")
	}

	// Leader proposes a write
	entry, err := node2.Propose("CREATE", "/app", []byte("hello"))
	if err != nil {
		t.Fatalf("leader should accept proposal: %v", err)
	}

	if entry.TxID != 1 {
		t.Fatalf("expected TxID 1, got %d", entry.TxID)
	}
	if entry.Path != "/app" {
		t.Fatalf("expected path /app, got %s", entry.Path)
	}
}

func TestPropose_FollowerRejects(t *testing.T) {
	node := newTestNode("node-1")

	// node-1 is a follower — should reject
	_, err := node.Propose("CREATE", "/app", []byte("hello"))
	if err == nil {
		t.Fatal("follower should reject proposal")
	}
}

// --- Commit tests ---

func TestCommit_AdvancesAfterMajority(t *testing.T) {
	nodes, _ := newTestCluster()
	node2 := nodes["node-2"]

	// Make node-2 the leader.
	voteReq := node2.StartElection()
	votes := 1
	for _, peer := range node2.config.OtherPeers() {
		resp := nodes[peer.ID].HandleRequestVote(voteReq)
		node2.CollectVote(resp, &votes)
	}

	// Propose 3 entries.
	node2.Propose("CREATE", "/app", []byte("v1"))
	node2.Propose("SET", "/app", []byte("v2"))
	node2.Propose("CREATE", "/db", []byte("v3"))

	// Before replication: commitIndex should be 0.
	if ci := node2.GetCommitIndex(); ci != 0 {
		t.Fatalf("expected commitIndex=0 before replication, got %d", ci)
	}

	// One leaderTick: sends entries to followers + checks majority.
	node2.leaderTick()

	// After replication: all 3 nodes have all 3 entries.
	// Quorum = 2. Leader has them (1) + each follower confirmed (2).
	// So commitIndex should advance to 3.
	if ci := node2.GetCommitIndex(); ci != 3 {
		t.Fatalf("expected commitIndex=3 after replication, got %d", ci)
	}
}

func TestCommit_FollowerLearnsCommitIndex(t *testing.T) {
	nodes, _ := newTestCluster()
	node2 := nodes["node-2"]

	// Make node-2 the leader.
	voteReq := node2.StartElection()
	votes := 1
	for _, peer := range node2.config.OtherPeers() {
		resp := nodes[peer.ID].HandleRequestVote(voteReq)
		node2.CollectVote(resp, &votes)
	}

	// Propose and replicate.
	node2.Propose("CREATE", "/app", []byte("v1"))
	node2.Propose("SET", "/app", []byte("v2"))
	node2.leaderTick() // sends entries + advances commitIndex

	// Leader's commitIndex should be 2.
	if ci := node2.GetCommitIndex(); ci != 2 {
		t.Fatalf("leader commitIndex should be 2, got %d", ci)
	}

	// Followers don't know yet — commitIndex was sent BEFORE it advanced.
	// The next leaderTick will carry the updated commitIndex.
	node2.leaderTick()

	// Now followers should have commitIndex = 2.
	for _, id := range []NodeID{"node-1", "node-3"} {
		if ci := nodes[id].GetCommitIndex(); ci != 2 {
			t.Fatalf("%s commitIndex should be 2, got %d", id, ci)
		}
	}
}

// --- Replication tests ---

func TestReplication_LeaderSendsEntriesToFollowers(t *testing.T) {
	nodes, _ := newTestCluster()
	node2 := nodes["node-2"]

	// Make node-2 the leader.
	voteReq := node2.StartElection()
	votes := 1
	for _, peer := range node2.config.OtherPeers() {
		resp := nodes[peer.ID].HandleRequestVote(voteReq)
		node2.CollectVote(resp, &votes)
	}

	if node2.GetState().Role != Leader {
		t.Fatal("node-2 should be leader")
	}

	// Leader proposes 3 entries.
	node2.Propose("CREATE", "/app", []byte("v1"))
	node2.Propose("SET", "/app", []byte("v2"))
	node2.Propose("CREATE", "/db", []byte("v3"))

	// Leader's log should have 3 entries.
	node2.mu.Lock()
	if len(node2.log) != 3 {
		t.Fatalf("leader should have 3 entries, got %d", len(node2.log))
	}
	node2.mu.Unlock()

	// Followers have nothing yet — entries haven't been sent.
	nodes["node-1"].mu.Lock()
	if len(nodes["node-1"].log) != 0 {
		t.Fatal("node-1 should have 0 entries before replication")
	}
	nodes["node-1"].mu.Unlock()

	// Trigger one leaderTick — this sends entries to followers.
	node2.leaderTick()

	// Now both followers should have all 3 entries.
	for _, id := range []NodeID{"node-1", "node-3"} {
		node := nodes[id]
		node.mu.Lock()
		count := len(node.log)
		lastTx := node.lastLogTxID
		node.mu.Unlock()

		if count != 3 {
			t.Fatalf("%s should have 3 entries, got %d", id, count)
		}
		if lastTx != 3 {
			t.Fatalf("%s should have lastLogTxID=3, got %d", id, lastTx)
		}
	}

	// Leader's nextIndex should have advanced for both peers.
	node2.mu.Lock()
	for _, peer := range node2.config.OtherPeers() {
		if node2.nextIndex[peer.ID] != 4 {
			t.Fatalf("nextIndex[%s] should be 4, got %d", peer.ID, node2.nextIndex[peer.ID])
		}
		if node2.matchIndex[peer.ID] != 3 {
			t.Fatalf("matchIndex[%s] should be 3, got %d", peer.ID, node2.matchIndex[peer.ID])
		}
	}
	node2.mu.Unlock()
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

// --- Election tests ---

// This test simulates a full election with 3 nodes.
// No network — just direct method calls.
// Read it top to bottom to see exactly how an election works.
func TestElection_FullFlow(t *testing.T) {
	// Setup: 3 nodes, all start as followers in term 0.
	node1 := newTestNode("node-1")
	node2 := newTestNode("node-2")
	node3 := newTestNode("node-3")

	// --- BEFORE ELECTION ---
	// All three are followers. No leader exists.
	s1 := node1.GetState()
	if s1.Role != Follower || s1.CurrentTerm != 0 {
		t.Fatal("node-1 should be follower in term 0")
	}

	// --- node-2 STARTS AN ELECTION ---
	// Imagine node-2's heartbeat timer expired (no leader heartbeat received).
	// node-2 calls StartElection():
	//   1. term goes from 0 → 1
	//   2. becomes candidate
	//   3. votes for itself
	//   4. returns a vote request to send to others
	voteReq := node2.StartElection()

	s2 := node2.GetState()
	if s2.Role != Candidate {
		t.Fatalf("node-2 should be candidate, got %s", s2.Role)
	}
	if s2.CurrentTerm != 1 {
		t.Fatalf("node-2 should be in term 1, got %d", s2.CurrentTerm)
	}
	if s2.VotedFor != "node-2" {
		t.Fatalf("node-2 should have voted for itself, got %s", s2.VotedFor)
	}

	// --- node-2 SENDS vote request to node-1 and node-3 ---
	// In real life this goes over the network.
	// In this test we just call the method directly.
	resp1 := node1.HandleRequestVote(voteReq)
	resp3 := node3.HandleRequestVote(voteReq)

	// Both should grant their vote (term 1 is new, no one voted yet)
	if !resp1.VoteGranted {
		t.Fatal("node-1 should vote for node-2")
	}
	if !resp3.VoteGranted {
		t.Fatal("node-3 should vote for node-2")
	}

	// --- node-2 COLLECTS votes ---
	// It already has 1 vote (itself). Quorum for 3 nodes = 2.
	votes := 1 // self-vote

	won := node2.CollectVote(resp1, &votes)
	// votes is now 2, quorum is 2 → should win
	if !won {
		t.Fatal("node-2 should win after getting node-1's vote")
	}

	// --- AFTER ELECTION ---
	// node-2 is now leader
	s2 = node2.GetState()
	if s2.Role != Leader {
		t.Fatalf("node-2 should be leader, got %s", s2.Role)
	}
	if s2.LeaderID != "node-2" {
		t.Fatalf("node-2 should know it's the leader, got %s", s2.LeaderID)
	}

	// node-1 and node-3 are followers who voted for node-2
	s1 = node1.GetState()
	if s1.VotedFor != "node-2" {
		t.Fatalf("node-1 should have voted for node-2, got %s", s1.VotedFor)
	}
}

// This test shows what happens when an election fails (split vote).
func TestElection_SplitVote(t *testing.T) {
	node1 := newTestNode("node-1")
	node2 := newTestNode("node-2")
	node3 := newTestNode("node-3")

	// node-2 AND node-3 both start elections at the same time.
	// Both increment to term 1 and vote for themselves.
	voteReq2 := node2.StartElection()
	voteReq3 := node3.StartElection()

	// node-1 receives node-2's request first → votes for node-2
	resp := node1.HandleRequestVote(voteReq2)
	if !resp.VoteGranted {
		t.Fatal("node-1 should vote for node-2 (first request)")
	}

	// node-1 then receives node-3's request → rejects (already voted)
	resp = node1.HandleRequestVote(voteReq3)
	if resp.VoteGranted {
		t.Fatal("node-1 should reject node-3 (already voted for node-2)")
	}

	// node-2 asks node-3 → node-3 already voted for itself → rejects
	resp = node3.HandleRequestVote(voteReq2)
	if resp.VoteGranted {
		t.Fatal("node-3 should reject node-2 (voted for itself)")
	}

	// Result:
	//   node-2 has 2 votes: itself + node-1       → wins (quorum = 2)
	//   node-3 has 1 vote:  itself                 → loses
	votes2 := 1 // self-vote
	won := node2.CollectVote(
		RequestVoteResponse{Term: 1, VoteGranted: true}, // node-1's yes
		&votes2,
	)
	if !won {
		t.Fatal("node-2 should win with 2 votes")
	}

	votes3 := 1 // self-vote
	won = node3.CollectVote(
		RequestVoteResponse{Term: 1, VoteGranted: false}, // node-1's no
		&votes3,
	)
	if won {
		t.Fatal("node-3 should not win with only 1 vote")
	}
}

// --- Automatic election test ---

// This test starts 3 connected nodes, starts the loop, and waits
// for a leader to be elected automatically. No manual calls.
func TestElection_Automatic(t *testing.T) {
	nodes, _ := newTestCluster()

	// Make timeouts very short so the test runs fast.
	for _, node := range nodes {
		node.heartbeatInterval = 20 * time.Millisecond
		node.electionTimeoutMin = 50 * time.Millisecond
		node.electionTimeoutMax = 150 * time.Millisecond
	}

	// Start all 3 nodes. Each runs its loop in a goroutine.
	for _, node := range nodes {
		node.Run()
	}

	// Wait for an election to happen (should take < 200ms)
	time.Sleep(500 * time.Millisecond)

	// Stop all nodes
	for _, node := range nodes {
		node.Stop()
	}

	// Check: exactly one node should be leader
	leaderCount := 0
	var leaderID NodeID
	for id, node := range nodes {
		state := node.GetState()
		if state.Role == Leader {
			leaderCount++
			leaderID = id
		}
	}

	if leaderCount != 1 {
		t.Fatalf("expected exactly 1 leader, got %d", leaderCount)
	}

	// Check: all followers should know who the leader is
	for id, node := range nodes {
		if id == leaderID {
			continue
		}
		state := node.GetState()
		if state.LeaderID != leaderID {
			t.Fatalf("node %s thinks leader is %s, but actual leader is %s",
				id, state.LeaderID, leaderID)
		}
	}

	t.Logf("leader elected: %s", leaderID)
}
