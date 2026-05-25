package cluster

import (
	"fmt"
	"testing"
	"time"

	"github.com/syamsularifin/zookeeper/internal/wal"
	"github.com/syamsularifin/zookeeper/internal/znode"
)

// memoryStorage is an in-memory implementation of Storage for tests.
// No temp files, no disk I/O — just slices.
type memoryStorage struct {
	entries []wal.Entry
	applied []wal.Entry
	tree    *znode.DataTree // optional, for integration tests
}

func newMemoryStorage() *memoryStorage {
	return &memoryStorage{}
}

func (ms *memoryStorage) AppendWAL(entry wal.Entry) error {
	ms.entries = append(ms.entries, entry)
	return nil
}

func (ms *memoryStorage) ApplyTree(entry wal.Entry) error {
	ms.applied = append(ms.applied, entry)
	if ms.tree != nil {
		switch entry.Op {
		case "CREATE":
			return ms.tree.Create(entry.Path, entry.Data)
		case "SET":
			return ms.tree.Set(entry.Path, entry.Data)
		case "DELETE":
			return ms.tree.Delete(entry.Path)
		}
	}
	return nil
}

func (ms *memoryStorage) GetWALEntriesFrom(fromTxID int64) []wal.Entry {
	idx := int(fromTxID - 1)
	if idx < 0 || idx >= len(ms.entries) {
		return nil
	}
	return ms.entries[idx:]
}

func (ms *memoryStorage) LastWALTxID() int64 {
	if len(ms.entries) == 0 {
		return 0
	}
	return ms.entries[len(ms.entries)-1].TxID
}

func (ms *memoryStorage) TruncateWALFrom(fromTxID int64) {
	idx := int(fromTxID - 1)
	if idx < 0 {
		idx = 0
	}
	if idx < len(ms.entries) {
		ms.entries = ms.entries[:idx]
	}
}

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

// failingTransport simulates all peers being unreachable.
type failingTransport struct{}

func (ft *failingTransport) SendRequestVote(peer Peer, req RequestVoteRequest) (RequestVoteResponse, error) {
	return RequestVoteResponse{}, fmt.Errorf("peer %s unreachable", peer.ID)
}

func (ft *failingTransport) SendAppendEntries(peer Peer, req AppendEntriesRequest) (AppendEntriesResponse, error) {
	return AppendEntriesResponse{}, fmt.Errorf("peer %s unreachable", peer.ID)
}

var testPeers = []Peer{
	{ID: "node-1", Addr: "localhost:3001"},
	{ID: "node-2", Addr: "localhost:3002"},
	{ID: "node-3", Addr: "localhost:3003"},
}

func newTestNode(id NodeID) (*RaftNode, *memoryStorage) {
	ms := newMemoryStorage()
	node := NewRaftNode(Config{Self: id, Peers: testPeers}, &fakeTransport{}, ms)
	return node, ms
}

// newTestCluster creates 3 connected nodes that can talk to each other.
func newTestCluster() (map[NodeID]*RaftNode, map[NodeID]*memoryStorage) {
	ft := &fakeTransport{nodes: make(map[NodeID]*RaftNode)}
	stores := make(map[NodeID]*memoryStorage)

	for _, p := range testPeers {
		ms := newMemoryStorage()
		stores[p.ID] = ms
		node := NewRaftNode(Config{Self: p.ID, Peers: testPeers}, ft, ms)
		ft.nodes[p.ID] = node
	}

	return ft.nodes, stores
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

	// Leader appends to WAL (appendEntry = no commit wait)
	entry, err := node2.appendEntry("CREATE", "/app", []byte("hello"))
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
	node, _ := newTestNode("node-1")

	// node-1 is a follower — should reject
	_, err := node.appendEntry("CREATE", "/app", []byte("hello"))
	if err == nil {
		t.Fatal("follower should reject proposal")
	}
}

// TestPropose_SynchronousCommit proves that Propose replicates immediately
// and returns only after the entry is committed (majority confirmed).
//
// No tick loop needed — Propose calls leaderTick internally.
func TestPropose_SynchronousCommit(t *testing.T) {
	nodes, stores := newTestCluster()
	node2 := nodes["node-2"]

	// Give each node a tree so apply works.
	for _, ms := range stores {
		ms.tree = znode.NewDataTree()
	}

	// Elect node-2 as leader.
	voteReq := node2.StartElection()
	votes := 1
	for _, peer := range node2.config.OtherPeers() {
		resp := nodes[peer.ID].HandleRequestVote(voteReq)
		node2.CollectVote(resp, &votes)
	}

	// Propose replicates immediately — no Run() needed.
	entry, err := node2.Propose("CREATE", "/app", []byte("hello"))
	if err != nil {
		t.Fatalf("Propose should succeed: %v", err)
	}

	if entry.TxID != 1 {
		t.Fatalf("expected TxID 1, got %d", entry.TxID)
	}

	// After Propose returns, the entry is committed AND applied.
	// Leader's tree should have the data.
	data, err := stores["node-2"].tree.Get("/app")
	if err != nil {
		t.Fatalf("leader tree should have /app: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected 'hello', got '%s'", string(data))
	}
}

// TestPropose_FailsWithoutConsensus proves that Propose returns an error
// when majority of followers are unreachable.
//
// In a 3-node cluster, quorum = 2. If both followers are down,
// only the leader has the entry → no majority → no commit → error.
func TestPropose_FailsWithoutConsensus(t *testing.T) {
	// Create a leader with a transport where all peers fail.
	failTransport := &failingTransport{}
	ms := newMemoryStorage()
	node := NewRaftNode(Config{Self: "node-1", Peers: testPeers}, failTransport, ms)

	// Force node to be leader.
	node.mu.Lock()
	node.state.Role = Leader
	node.state.LeaderID = "node-1"
	node.nextIndex = map[NodeID]int64{"node-2": 1, "node-3": 1}
	node.matchIndex = map[NodeID]int64{"node-2": 0, "node-3": 0}
	node.mu.Unlock()

	// Propose should fail — no followers reachable, no consensus.
	_, err := node.Propose("CREATE", "/app", []byte("hello"))
	if err == nil {
		t.Fatal("Propose should fail without consensus")
	}
	t.Logf("got expected error: %v", err)
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
	node2.appendEntry("CREATE", "/app", []byte("v1"))
	node2.appendEntry("SET", "/app", []byte("v2"))
	node2.appendEntry("CREATE", "/db", []byte("v3"))
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
	node2.appendEntry("CREATE", "/app", []byte("v1"))
	node2.appendEntry("SET", "/app", []byte("v2"))
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

// --- Apply tests ---

func TestApply_LeaderAppliesCommittedEntries(t *testing.T) {
	nodes, stores := newTestCluster()
	node2 := nodes["node-2"]

	// Make node-2 the leader.
	voteReq := node2.StartElection()
	votes := 1
	for _, peer := range node2.config.OtherPeers() {
		resp := nodes[peer.ID].HandleRequestVote(voteReq)
		node2.CollectVote(resp, &votes)
	}

	// Propose 2 entries and replicate.
	node2.appendEntry("CREATE", "/app", []byte("v1"))
	node2.appendEntry("SET", "/app", []byte("v2"))
	node2.leaderTick()

	// Leader should have applied both entries.
	applied := stores["node-2"].applied
	if len(applied) != 2 {
		t.Fatalf("expected 2 applied entries, got %d", len(applied))
	}
	if applied[0].Path != "/app" || applied[0].Op != "CREATE" {
		t.Fatalf("first applied entry wrong: %+v", applied[0])
	}
	if applied[1].Path != "/app" || applied[1].Op != "SET" {
		t.Fatalf("second applied entry wrong: %+v", applied[1])
	}
}

func TestApply_FollowerAppliesAfterLearningCommitIndex(t *testing.T) {
	nodes, stores := newTestCluster()
	node2 := nodes["node-2"]

	// Make node-2 the leader.
	voteReq := node2.StartElection()
	votes := 1
	for _, peer := range node2.config.OtherPeers() {
		resp := nodes[peer.ID].HandleRequestVote(voteReq)
		node2.CollectVote(resp, &votes)
	}

	// Propose and replicate.
	node2.appendEntry("CREATE", "/app", []byte("v1"))
	node2.leaderTick() // sends entries + advances commitIndex

	// Follower hasn't applied yet — it learned commitIndex=0 in this tick.
	if len(stores["node-1"].applied) != 0 {
		t.Fatalf("expected 0 applied on follower after first tick, got %d", len(stores["node-1"].applied))
	}

	// Second tick carries the updated commitIndex.
	node2.leaderTick()

	// Now follower should have applied the entry.
	applied := stores["node-1"].applied
	if len(applied) != 1 {
		t.Fatalf("expected 1 applied entry on follower, got %d", len(applied))
	}
	if applied[0].Path != "/app" {
		t.Fatalf("applied entry wrong: %+v", applied[0])
	}
}

// --- Replication tests ---

func TestReplication_LeaderSendsEntriesToFollowers(t *testing.T) {
	nodes, stores := newTestCluster()
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
	node2.appendEntry("CREATE", "/app", []byte("v1"))
	node2.appendEntry("SET", "/app", []byte("v2"))
	node2.appendEntry("CREATE", "/db", []byte("v3"))

	// Leader's WAL should have 3 entries.
	if len(stores["node-2"].entries) != 3 {
		t.Fatalf("leader should have 3 entries, got %d", len(stores["node-2"].entries))
	}

	// Followers have nothing yet — entries haven't been sent.
	if len(stores["node-1"].entries) != 0 {
		t.Fatal("node-1 should have 0 entries before replication")
	}

	// Trigger one leaderTick — this sends entries to followers.
	node2.leaderTick()

	// Now both followers should have all 3 entries in their WAL.
	for _, id := range []NodeID{"node-1", "node-3"} {
		count := len(stores[id].entries)
		lastTx := stores[id].LastWALTxID()

		if count != 3 {
			t.Fatalf("%s should have 3 entries, got %d", id, count)
		}
		if lastTx != 3 {
			t.Fatalf("%s should have lastWALTxID=3, got %d", id, lastTx)
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
	node, _ := newTestNode("node-1")

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
	node, _ := newTestNode("node-1")

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

// TestAppendEntries_TruncatesConflictingLog proves that when a follower
// has an orphaned entry from a different term, the new leader's entries
// overwrite it via the prevLog consistency check.
//
// Scenario:
//   1. Follower has entries [1, 2, 3, 4, 5_old]  (5_old from old leader, term 1)
//   2. New leader (term 2) sends entries starting at 5, with prevLog = entry 4
//   3. prevLog matches (entry 4 is same term) → accepted
//   4. Truncate from TxID=5, append leader's new entries
//   5. Result: [1, 2, 3, 4, 5_new, 6]
func TestAppendEntries_TruncatesConflictingLog(t *testing.T) {
	ms := newMemoryStorage()
	node := NewRaftNode(Config{Self: "node-1", Peers: testPeers}, &fakeTransport{}, ms)

	// Simulate follower having entries 1-5, where entry 5 is from old leader (term 1).
	ms.entries = []wal.Entry{
		{TxID: 1, Term: 1, Op: "CREATE", Path: "/a"},
		{TxID: 2, Term: 1, Op: "CREATE", Path: "/b"},
		{TxID: 3, Term: 1, Op: "CREATE", Path: "/c"},
		{TxID: 4, Term: 1, Op: "CREATE", Path: "/d"},
		{TxID: 5, Term: 1, Op: "CREATE", Path: "/old"}, // ← orphaned, will be overwritten
	}

	// New leader (term 2) sends entries starting at 5.
	// PrevLog = entry 4, which matches (term 1).
	resp := node.HandleAppendEntries(AppendEntriesRequest{
		Term:        2,
		LeaderID:    "node-2",
		PrevLogTxID: 4,
		PrevLogTerm: 1,
		Entries: []wal.Entry{
			{TxID: 5, Term: 2, Op: "CREATE", Path: "/new", Data: []byte("correct")},
			{TxID: 6, Term: 2, Op: "CREATE", Path: "/e", Data: []byte("v6")},
		},
	})

	if !resp.Success {
		t.Fatal("should accept AppendEntries (prevLog matches)")
	}

	// Follower should have 6 entries: 1-4 unchanged, 5 replaced, 6 new.
	if len(ms.entries) != 6 {
		t.Fatalf("expected 6 entries, got %d", len(ms.entries))
	}

	// Entry 5 should be the leader's version (term 2), not the orphaned one.
	if ms.entries[4].Path != "/new" || ms.entries[4].Term != 2 {
		t.Fatalf("entry 5 wrong: %+v", ms.entries[4])
	}

	// Entry 6 should be appended.
	if ms.entries[5].Path != "/e" {
		t.Fatalf("entry 6 wrong: %+v", ms.entries[5])
	}
}

// TestAppendEntries_RejectsPrevLogMismatch proves that when the follower's
// entry at PrevLogTxID has a different term, it rejects and truncates.
// The leader will back up and retry with an earlier prevLog.
func TestAppendEntries_RejectsPrevLogMismatch(t *testing.T) {
	ms := newMemoryStorage()
	node := NewRaftNode(Config{Self: "node-1", Peers: testPeers}, &fakeTransport{}, ms)

	// Follower has entries 1-3, where entry 3 is from term 1.
	ms.entries = []wal.Entry{
		{TxID: 1, Term: 1, Op: "CREATE", Path: "/a"},
		{TxID: 2, Term: 1, Op: "CREATE", Path: "/b"},
		{TxID: 3, Term: 1, Op: "CREATE", Path: "/c"}, // term 1
	}

	// Leader says prevLog should be TxID=3, Term=2. But follower has term=1 at TxID=3.
	resp := node.HandleAppendEntries(AppendEntriesRequest{
		Term:        3,
		LeaderID:    "node-2",
		PrevLogTxID: 3,
		PrevLogTerm: 2, // mismatch! follower has term 1
		Entries: []wal.Entry{
			{TxID: 4, Term: 3, Op: "CREATE", Path: "/d"},
		},
	})

	if resp.Success {
		t.Fatal("should reject: prevLog term mismatch")
	}

	// Follower should have truncated entry 3 (the conflicting one).
	if len(ms.entries) != 2 {
		t.Fatalf("expected 2 entries after truncation, got %d", len(ms.entries))
	}

	// LastLogTxID should be 2 (after truncating entry 3).
	if resp.LastLogTxID != 2 {
		t.Fatalf("expected LastLogTxID=2, got %d", resp.LastLogTxID)
	}
}

// --- RequestVote tests ---

func TestRequestVote_GrantVote(t *testing.T) {
	node, _ := newTestNode("node-1")

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
	node, _ := newTestNode("node-1")

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
	node, _ := newTestNode("node-1")

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
	node, _ := newTestNode("node-1")

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
	// Pre-populate storage with 10 entries so the node has a long log.
	ms := newMemoryStorage()
	for i := int64(1); i <= 10; i++ {
		ms.entries = append(ms.entries, wal.Entry{TxID: i})
	}
	node := NewRaftNode(Config{Self: "node-1", Peers: testPeers}, &fakeTransport{}, ms)

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
	node1, _ := newTestNode("node-1")
	node2, _ := newTestNode("node-2")
	node3, _ := newTestNode("node-3")

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
	node1, _ := newTestNode("node-1")
	node2, _ := newTestNode("node-2")
	node3, _ := newTestNode("node-3")

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

// --- Integration test: Raft → Tree ---

// This test proves the full flow end-to-end:
//
//	propose on leader → replicate → commit → apply → all trees have the data
//
// Each node has its own DataTree inside its memoryStorage.
// After the flow completes, all 3 trees should be identical.
func TestIntegration_RaftToTree(t *testing.T) {
	nodes, stores := newTestCluster()

	// Give each node a tree in its storage.
	for _, ms := range stores {
		ms.tree = znode.NewDataTree()
	}

	// Elect node-2 as leader.
	node2 := nodes["node-2"]
	voteReq := node2.StartElection()
	votes := 1
	for _, peer := range node2.config.OtherPeers() {
		resp := nodes[peer.ID].HandleRequestVote(voteReq)
		node2.CollectVote(resp, &votes)
	}

	if node2.GetState().Role != Leader {
		t.Fatal("node-2 should be leader")
	}

	// Client writes to the leader.
	node2.appendEntry("CREATE", "/app", []byte("hello"))
	node2.appendEntry("CREATE", "/app/config", []byte("v1"))

	// Tick 1: replicate entries + leader commits + leader applies.
	node2.leaderTick()

	// Leader's tree should have the data now.
	data, err := stores["node-2"].tree.Get("/app")
	if err != nil {
		t.Fatalf("leader tree should have /app: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected 'hello', got '%s'", string(data))
	}

	// Tick 2: followers learn commitIndex + followers apply.
	node2.leaderTick()

	// All 3 trees should have the same data.
	for id, ms := range stores {
		data, err := ms.tree.Get("/app")
		if err != nil {
			t.Fatalf("%s tree should have /app: %v", id, err)
		}
		if string(data) != "hello" {
			t.Fatalf("%s: expected 'hello', got '%s'", id, string(data))
		}

		data, err = ms.tree.Get("/app/config")
		if err != nil {
			t.Fatalf("%s tree should have /app/config: %v", id, err)
		}
		if string(data) != "v1" {
			t.Fatalf("%s: expected 'v1', got '%s'", id, string(data))
		}
	}
}
