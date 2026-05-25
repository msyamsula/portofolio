# 06 - Raft Consensus

## The Problem

In Phase 1, we built a single-node ZooKeeper. It works, but if that one node dies, all data is gone (or at least unavailable until restart). In a production system, clients need the service to keep running even when machines crash.

The answer: replicate the data across multiple nodes. If one dies, the others still have everything.

But now we have a new problem: how do the nodes agree on what the data is? If two clients write at the same time, which write wins? If a node was disconnected for a while, how does it catch up? This is the **consensus problem**, and Raft is an algorithm that solves it.

## What We Built

We implemented Raft's core consensus algorithm. Every write goes through the leader, gets replicated to a majority of nodes, and only then is considered committed.

### The Architecture

```
              Client (zkcli)
                   │
                   ▼
              ┌─────────┐
              │  Leader  │  ← only node that accepts writes
              │ RaftNode │
              │  +Store  │
              └─────────┘
               /         \
              ▼           ▼
        ┌──────────┐  ┌──────────┐
        │ Follower  │  │ Follower  │
        │ RaftNode  │  │ RaftNode  │
        │  +Store   │  │  +Store   │
        └──────────┘  └──────────┘
```

Each node in the cluster is a `RaftNode` that owns a `Store`. The `Store` has the WAL (disk) and DataTree (memory) from Phase 1. The `RaftNode` coordinates replication.

### Key Design Decision: RaftNode Owns Store

RaftNode is the unit of deployment (one per container). It contains a Store, not the other way around. In cluster mode, all writes go through `RaftNode.Propose()`, which replicates first, then writes to the Store.

## The Three Roles

Every node is in exactly one state at any time:

```
Follower ──── (no heartbeat for 300-500ms) ────→ Candidate
Candidate ─── (gets majority votes) ──────────→ Leader
Candidate ─── (someone else wins) ────────────→ Follower
Leader ─────── (discovers higher term) ───────→ Follower
```

- **Follower**: "I follow the leader. I replicate what I'm told."
- **Candidate**: "I think the leader is dead. I'm running for election."
- **Leader**: "I handle all writes and send heartbeats."

## Terms

Time in Raft is divided into **terms** (like election cycles). Terms only go up, never down. Every message carries a term number. If a node sees a higher term, it knows it's out of date and steps down to follower.

```
Term 1: node-1 is leader
Term 2: node-1 dies → node-3 wins election → becomes leader
Term 3: node-3 dies → node-2 wins election → becomes leader
```

## Elections

When a follower doesn't hear from the leader for 300-500ms (random per node to prevent split votes), it starts an election:

1. Increment term
2. Become candidate
3. Vote for itself
4. Ask all other nodes for votes

A node grants a vote if:
- The candidate's term >= its own term
- It hasn't already voted this term
- The candidate's log is at least as up-to-date (prevents electing a node that's missing committed entries)

Quorum = majority. In a 3-node cluster, quorum = 2. In a 5-node cluster, quorum = 3.

## The Write Path (Propose)

This is the most important flow. When a client writes:

```
Client: "Create /app hello"
    │
    ▼
Leader.Propose(CREATE, "/app", "hello")
    │
    ├── 1. Create entry IN MEMORY ONLY (TxID=N, Term=T)
    │      NOT written to leader's WAL yet.
    │
    ├── 2. Send AppendEntries to all followers
    │      Each follower writes to its WAL and responds Success=true
    │
    ├── 3. Count responses. Majority confirmed?
    │      │
    │      ├── NO → return error. Nothing was written on leader. Clean failure.
    │      │
    │      └── YES → continue...
    │
    ├── 4. Write to leader's WAL (now safe — majority has it)
    │
    ├── 5. Apply to leader's DataTree
    │
    └── 6. Return success to client
```

### Why "Consensus-First" (WAL After Consensus)?

If the leader wrote to its WAL before consensus and then consensus failed, the entry would still be in the leader's WAL. The background tick loop would eventually commit it — but the client was told it failed. That's a lie.

By writing the leader's WAL AFTER consensus:
- Error truly means "nothing happened on the leader"
- No ghost entries that get committed later
- Client can safely retry

## Log Consistency (PrevLog Check)

A critical correctness concern: what if a follower has stale/conflicting entries from a previous leader?

### The Problem

```
Leader (term 1) writes entries 1-5 to follower-A
Leader crashes before replicating entry 5 to follower-B
New leader (term 2) starts writing entry 5 (different content)

Now follower-A has entry 5 from term 1 (orphaned)
New leader has entry 5 from term 2 (correct)
```

### The Solution: O(1) PrevLog Check

Every `AppendEntriesRequest` includes `PrevLogTxID` and `PrevLogTerm` — the TxID and Term of the entry immediately BEFORE the batch.

The follower checks ONE entry:
- **Match**: logs agree up to this point. Accept the new entries.
- **Missing**: follower's log is shorter. Reject — leader backs up.
- **Wrong term**: conflict detected. Truncate from this point. Reject — leader will resend.

This is O(1) per request, not O(entries). The Raft paper calls this the "consistency check" (section 5.3).

### Why Term on Entries?

Each WAL entry now carries a `Term` field (which Raft term it was created in). Two entries at the same TxID but different Terms mean they were created by different leaders — a conflict. Without Term, we couldn't detect this.

```go
type Entry struct {
    TxID int64  `json:"tx_id"`
    Term int64  `json:"term"`    // ← added for Raft
    Op   OpType `json:"op"`
    Path string `json:"path"`
    Data []byte `json:"data,omitempty"`
}
```

## Heartbeats and Replication (leaderTick)

The leader runs a tick loop every 50ms. On each tick:

1. For each follower, look up `nextIndex[peer]` — where is this peer in the log?
2. Grab entries from `nextIndex` onward from the in-memory cache
3. Compute `PrevLogTxID` and `PrevLogTerm` for consistency check
4. Send `AppendEntries` (entries + prevLog + commitIndex)
5. On success → advance `nextIndex` and `matchIndex`
6. On failure → jump `nextIndex` to `resp.LastLogTxID + 1`
7. After all peers, call `advanceCommitIndex()`

A heartbeat is just an `AppendEntries` with empty `Entries`. Same message, same logic.

## Commit Index and Apply

- **commitIndex**: highest TxID replicated to a majority. Only the leader advances this.
- **lastApplied**: highest TxID applied to the DataTree. Always <= commitIndex.
- Followers learn commitIndex from the leader via `AppendEntries.LeaderCommitIndex`.

`advanceCommitIndex` collects all `matchIndex` values (including leader's own), sorts descending, picks the quorum-th value. That's the highest TxID a majority has.

`applyCommitted` closes the gap between `lastApplied` and `commitIndex` by applying entries to the DataTree.

## Storage Interface

RaftNode depends on a `Storage` interface, not a concrete `*store.Store`. This decouples the packages and makes testing easy:

```go
type Storage interface {
    AppendWAL(entry wal.Entry) error        // write to WAL + cache
    ApplyTree(entry wal.Entry) error        // apply to DataTree
    GetWALEntriesFrom(fromTxID int64) []wal.Entry  // read from cache
    LastWALTxID() int64                     // last entry TxID
    TruncateWALFrom(fromTxID int64)         // remove conflicting entries
}
```

- **Production**: `*store.Store` implements this. WAL goes to disk, cache in memory.
- **Tests**: `memoryStorage` implements it. Pure in-memory, no temp files.

## In-Memory WAL Cache

Store keeps `entries []wal.Entry` in memory, mirroring the disk WAL. This exists because the leader needs fast access to entries for replication (`GetWALEntriesFrom(nextIndex)`). Reading from disk every 50ms would be too slow.

The cache is populated during recovery (`replay()`) and appended to during `AppendWAL()`.

## Messages

Raft uses only TWO message types:

### 1. AppendEntries (leader → followers)

```go
type AppendEntriesRequest struct {
    Term              int64       // leader's current term
    LeaderID          NodeID      // who the leader is
    PrevLogTxID       int64       // TxID of entry before this batch
    PrevLogTerm       int64       // Term of that entry
    Entries           []wal.Entry // entries to replicate (empty = heartbeat)
    LeaderCommitIndex int64       // leader's commitIndex
}

type AppendEntriesResponse struct {
    Term        int64   // follower's term
    Success     bool    // accepted or rejected
    LastLogTxID int64   // follower's last entry (for fast catch-up)
}
```

### 2. RequestVote (candidate → all nodes)

```go
type RequestVoteRequest struct {
    Term        int64   // candidate's proposed term
    CandidateID NodeID  // who is asking
    LastLogTxID int64   // candidate's log length (for up-to-date check)
}

type RequestVoteResponse struct {
    Term        int64   // voter's term
    VoteGranted bool    // yes or no
}
```

## Test Coverage (24 tests across cluster package)

| Category | Tests | What they prove |
|----------|-------|-----------------|
| Propose | `TestPropose_LeaderAccepts`, `_FollowerRejects`, `_SynchronousCommit`, `_FailsWithoutConsensus` | Leader accepts writes, followers reject, synchronous commit, clean failure without consensus |
| Commit | `_AdvancesAfterMajority`, `_FollowerLearnsCommitIndex` | commitIndex advances after majority, followers learn it via heartbeat |
| Apply | `_LeaderAppliesCommittedEntries`, `_FollowerAppliesAfterLearningCommitIndex` | Entries applied to tree after commit, on both leader and follower |
| Replication | `_LeaderSendsEntriesToFollowers` | Entries replicated to all followers, nextIndex/matchIndex updated |
| AppendEntries | `_AcceptFromLeader`, `_RejectStaleTerm`, `_TruncatesConflictingLog`, `_RejectsPrevLogMismatch` | Accept valid, reject stale, truncate conflicts, detect term mismatches |
| RequestVote | `_GrantVote`, `_RejectStaleTerm`, `_RejectAlreadyVoted`, `_NewTermClearsVote`, `_RejectCandidateWithShorterLog` | Vote granting, rejection for all correct reasons |
| Election | `_FullFlow`, `_SplitVote`, `_Automatic` | Manual election, split vote handling, automatic election via tick loop |
| Integration | `_RaftToTree` | Full flow: propose → replicate → commit → apply → all trees match |

## Files

- `internal/cluster/raft.go` — RaftNode: core state machine, Propose, HandleAppendEntries, HandleRequestVote, elections, tick loop
- `internal/cluster/raft_test.go` — 24 tests with memoryStorage and fakeTransport
- `internal/cluster/message.go` — AppendEntries + RequestVote request/response structs
- `internal/cluster/config.go` — NodeID, Peer, Config, QuorumSize
- `internal/cluster/state.go` — Role (Follower/Candidate/Leader), NodeState
- `internal/cluster/transport.go` — Transport interface
- `internal/wal/wal.go` — Entry struct (with Term field), AppendEntry method
- `internal/store/store.go` — AppendWAL, ApplyTree, GetWALEntriesFrom, TruncateWALFrom, LastWALTxID
