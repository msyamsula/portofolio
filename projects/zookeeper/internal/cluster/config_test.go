package cluster

import "testing"

func TestQuorumSize(t *testing.T) {
	tests := []struct {
		nodes   int
		quorum  int
	}{
		{1, 1},  // single node: only need itself
		{3, 2},  // 3-node cluster: need 2
		{5, 3},  // 5-node cluster: need 3
		{7, 4},  // 7-node cluster: need 4
	}

	for _, tt := range tests {
		// Build a config with tt.nodes peers
		cfg := Config{Self: "node-1"}
		for i := 0; i < tt.nodes; i++ {
			cfg.Peers = append(cfg.Peers, Peer{ID: NodeID("node")})
		}

		got := cfg.QuorumSize()
		if got != tt.quorum {
			t.Errorf("%d nodes: expected quorum %d, got %d", tt.nodes, tt.quorum, got)
		}
	}
}

func TestOtherPeers(t *testing.T) {
	cfg := Config{
		Self: "node-2",
		Peers: []Peer{
			{ID: "node-1", Addr: "localhost:3001"},
			{ID: "node-2", Addr: "localhost:3002"},
			{ID: "node-3", Addr: "localhost:3003"},
		},
	}

	others := cfg.OtherPeers()

	if len(others) != 2 {
		t.Fatalf("expected 2 other peers, got %d", len(others))
	}

	// Self should not be in the list
	for _, p := range others {
		if p.ID == "node-2" {
			t.Fatal("self should not be in OtherPeers")
		}
	}
}
