package cluster

// Every node in a Raft cluster is in exactly ONE of three states:
//
//   Follower  → "I follow the leader. I don't make decisions."
//   Candidate → "I think the leader is dead. I'm running for election."
//   Leader    → "I'm in charge. I handle all writes and send heartbeats."
//
// State transitions:
//
//   All nodes start as Follower.
//
//   Follower ──── (no heartbeat for a while) ────→ Candidate
//   Candidate ─── (gets majority votes) ──────────→ Leader
//   Candidate ─── (someone else wins) ────────────→ Follower
//   Leader ─────── (discovers higher term) ───────→ Follower
//
// That's it. There's no "offline" or "joining" state.
// A node is always one of these three.

// Role represents the current state of a node.
type Role int

const (
	Follower  Role = iota // default starting state
	Candidate             // trying to become leader
	Leader                // handling writes, sending heartbeats
)

// String returns a human-readable name for the role.
func (r Role) String() string {
	switch r {
	case Follower:
		return "follower"
	case Candidate:
		return "candidate"
	case Leader:
		return "leader"
	default:
		return "unknown"
	}
}

// NodeState holds the Raft state for this node.
//
// In Raft, time is divided into "terms". A term is like an election cycle:
//
//   Term 1: node-1 is leader
//   Term 2: node-1 died, node-3 wins election, becomes leader
//   Term 3: node-3 died, node-2 wins election, becomes leader
//
// Terms only go up, never down. If a node sees a message with a higher
// term than its own, it immediately knows: "I'm out of date. I should
// step down to follower and accept the new term."
//
// This prevents stale leaders. If node-1 was leader in term 1, got
// network-partitioned, comes back, and sees term 3 messages, it knows
// it's no longer leader — even though it never received a "you're fired" message.
type NodeState struct {
	// Role is the current state: Follower, Candidate, or Leader.
	Role Role

	// CurrentTerm is the latest term this node has seen.
	// Starts at 0. Increases by 1 each time an election starts.
	//
	// Every Raft message carries a term number. If a node receives
	// a message with a higher term, it updates its own term and
	// steps down to follower. This is how stale leaders discover
	// they've been replaced.
	CurrentTerm int64

	// VotedFor records who this node voted for in the current term.
	// Empty string means "haven't voted yet this term".
	//
	// Rule: a node can vote for AT MOST one candidate per term.
	// This prevents split votes from producing two leaders.
	//
	// Example:
	//   Term 5 starts. node-2 asks for votes. node-1 votes for node-2.
	//   node-3 also asks for votes. node-1 says "no, I already voted for node-2."
	VotedFor NodeID

	// LeaderID is the ID of the current known leader.
	// Empty if no leader is known (e.g. during an election).
	// Followers use this to forward write requests to the leader.
	LeaderID NodeID
}

// NewNodeState creates the initial state: follower, term 0, no votes.
func NewNodeState() *NodeState {
	return &NodeState{
		Role:        Follower,
		CurrentTerm: 0,
		VotedFor:    "",
		LeaderID:    "",
	}
}
