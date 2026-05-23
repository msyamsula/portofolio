package cluster

// Raft uses only TWO types of messages for everything:
//
// 1. AppendEntries — sent by leader to followers.
//    - With entries: "here are new WAL entries, add them to your log"
//    - Without entries: "I'm still alive" (heartbeat)
//    A heartbeat is just an AppendEntries with an empty Entries list.
//
// 2. RequestVote — sent by candidate to all nodes.
//    "I want to be leader. Vote for me?"
//
// That's it. Two message types run the entire consensus algorithm.

import "github.com/syamsularifin/zookeeper/internal/wal"

// AppendEntriesRequest is sent by the leader to followers.
//
// When Entries is empty, this is a heartbeat — just saying "I'm alive."
// When Entries has items, the leader is replicating new WAL entries.
//
// The leader sends these periodically (e.g. every 150ms).
// If a follower doesn't receive one for a while, it assumes
// the leader is dead and starts an election.
type AppendEntriesRequest struct {
	// Term is the leader's current term.
	// If a follower sees a term higher than its own, it updates its term.
	// If a follower sees a term LOWER than its own, it rejects the message
	// — the sender is a stale leader that doesn't know it's been replaced.
	Term int64

	// LeaderID tells the follower who the leader is.
	// Followers use this to forward client write requests to the leader.
	LeaderID NodeID

	// Entries is the list of WAL entries to replicate.
	// Empty for heartbeats, non-empty for log replication.
	Entries []wal.Entry
}

// AppendEntriesResponse is the follower's reply to the leader.
type AppendEntriesResponse struct {
	// Term is the follower's current term.
	// If this is higher than the leader's term, the leader must step down.
	// This is how a stale leader discovers it's been replaced.
	Term int64

	// Success indicates if the follower accepted the entries.
	// false if the follower's term is higher (stale leader).
	Success bool

	// LastLogTxID tells the leader where this follower is in the log.
	// On rejection, the leader uses this to jump nextIndex directly
	// instead of decrementing by 1 each tick.
	//
	// Without this: leader guesses, backs up 1 per tick → slow.
	// With this:    leader jumps to LastLogTxID + 1 → one round trip.
	LastLogTxID int64
}

// RequestVoteRequest is sent by a candidate to all nodes during an election.
//
// The candidate says: "I'm starting an election for term X. Vote for me?"
type RequestVoteRequest struct {
	// Term is the new term the candidate is proposing.
	// Always = candidate's previous term + 1.
	// Example: leader was in term 3, leader dies, candidate starts term 4.
	Term int64

	// CandidateID is who is asking for votes.
	CandidateID NodeID

	// LastLogTxID is the TxID of the candidate's last WAL entry.
	// Voters use this to check: "is this candidate at least as up-to-date as me?"
	// A node won't vote for a candidate with a shorter log — that candidate
	// might be missing committed entries, and electing it would lose data.
	LastLogTxID int64
}

// RequestVoteResponse is the voter's reply to a candidate.
type RequestVoteResponse struct {
	// Term is the voter's current term.
	Term int64

	// VoteGranted is true if the voter voted for this candidate.
	// false if:
	//   - the voter already voted for someone else this term
	//   - the candidate's term is lower than the voter's
	//   - the candidate's log is less up-to-date than the voter's
	VoteGranted bool
}
