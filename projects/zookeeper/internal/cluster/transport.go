package cluster

// Transport is how a node sends messages to other nodes.
//
// Why an interface and not real networking code?
// Because we want to test the election logic WITHOUT starting real servers.
// In tests, we use a fake transport that just calls methods directly.
// In production, we'll have a gRPC transport that sends over the network.
//
//   Test:       node1.SendRequestVote → directly calls node2.HandleRequestVote
//   Production: node1.SendRequestVote → gRPC call to node2's server
//
// Same RaftNode code, different transport. This is the strategy pattern.
type Transport interface {
	// SendRequestVote sends a vote request to a peer and waits for the response.
	// Returns an error if the peer is unreachable (dead node, network issue).
	SendRequestVote(peer Peer, req RequestVoteRequest) (RequestVoteResponse, error)

	// SendAppendEntries sends a heartbeat (or log entries) to a peer.
	// Returns an error if the peer is unreachable.
	SendAppendEntries(peer Peer, req AppendEntriesRequest) (AppendEntriesResponse, error)
}
