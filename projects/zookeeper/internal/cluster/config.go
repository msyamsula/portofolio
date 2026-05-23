package cluster

// In Phase 1, we had one node. It didn't need to know about anyone else.
//
// In Phase 2, we have multiple nodes that must find each other.
// Each node needs to know:
//   - "Who am I?" (my own ID and address)
//   - "Who are the other nodes?" (their IDs and addresses)
//
// This is the simplest possible cluster config: a static list of nodes.
// Every node gets the same list at startup. No discovery, no DNS, no gossip.
// Just a hardcoded list.
//
// Example for a 3-node cluster:
//
//   Node 1 starts with: "I am node-1, the others are node-2 and node-3"
//   Node 2 starts with: "I am node-2, the others are node-1 and node-3"
//   Node 3 starts with: "I am node-3, the others are node-1 and node-2"
//
// All three have the SAME list of peers. They only differ in which ID is "me".

// NodeID is a unique identifier for a node in the cluster.
// We use a string so it's human-readable in logs: "node-1", "node-2", etc.
type NodeID string

// Peer represents one node in the cluster.
type Peer struct {
	// ID is the unique identifier for this node.
	ID NodeID

	// Addr is the network address for Raft communication (not client gRPC).
	// Example: "localhost:3001"
	//
	// Why a separate port from the client gRPC port?
	// Because client traffic (Create, Get, Set) and cluster traffic
	// (heartbeats, vote requests, log replication) are different concerns.
	// Keeping them on separate ports lets us manage them independently.
	Addr string
}

// Config holds the full cluster configuration.
type Config struct {
	// Self is this node's ID. Used to identify "which one am I?" in the peers list.
	Self NodeID

	// Peers is the complete list of ALL nodes in the cluster, including self.
	// Example for a 3-node cluster:
	//   [
	//     {ID: "node-1", Addr: "localhost:3001"},
	//     {ID: "node-2", Addr: "localhost:3002"},
	//     {ID: "node-3", Addr: "localhost:3003"},
	//   ]
	Peers []Peer
}

// OtherPeers returns all peers except self.
// Used when sending messages — you don't send heartbeats to yourself.
func (c *Config) OtherPeers() []Peer {
	var others []Peer
	for _, p := range c.Peers {
		if p.ID != c.Self {
			others = append(others, p)
		}
	}
	return others
}

// QuorumSize returns the number of nodes needed for a majority.
//
// In a 3-node cluster: quorum = 2 (majority of 3)
// In a 5-node cluster: quorum = 3 (majority of 5)
//
// Why majority? Because any two majorities overlap by at least one node.
// This guarantees that any decision made by a majority includes at least
// one node that knows about the previous decision.
//
// Example with 3 nodes:
//   Majority = any 2 of {A, B, C}
//   Possible majorities: {A,B}, {A,C}, {B,C}
//   Any two of these share at least one node.
//
// Formula: (n / 2) + 1
//   3 nodes → 2
//   5 nodes → 3
//   7 nodes → 4
//
// This is why clusters use odd numbers: 3, 5, 7.
// With 4 nodes, quorum is 3 — same as with 5 nodes, but you have
// one fewer node that can fail. 4 nodes is strictly worse than 5.
func (c *Config) QuorumSize() int {
	return len(c.Peers)/2 + 1
}
