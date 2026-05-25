# 07 - Known Issues and Roadmap to Production

## What Works Today

### Phase 1: Single Node (Complete)

| Component | Status | Files |
|-----------|--------|-------|
| ZNode tree (in-memory) | Done | `internal/znode/znode.go`, `tree.go` |
| WAL (write-ahead log) | Done | `internal/wal/wal.go` |
| Snapshots | Done | `internal/snapshot/snapshot.go` |
| Store (coordinator) | Done | `internal/store/store.go` |
| gRPC server | Done | `internal/server/server.go` |
| CLI client (zkcli) | Done | `cmd/zkcli/main.go` |
| Server binary (zknode) | Done | `cmd/zknode/main.go` |
| Recovery (snapshot + WAL replay) | Done | `store.New()` |

### Phase 2: Raft Consensus (In Progress)

| Component | Status | Files |
|-----------|--------|-------|
| Cluster config (static peers) | Done | `internal/cluster/config.go` |
| Node state (role, term, votedFor) | Done | `internal/cluster/state.go` |
| Raft messages (AppendEntries, RequestVote) | Done | `internal/cluster/message.go` |
| Transport interface | Done | `internal/cluster/transport.go` |
| Leader election | Done | `raft.go`: `StartElection`, `CollectVote`, `runElection` |
| Log replication (leader → followers) | Done | `raft.go`: `leaderTick` |
| Synchronous Propose (consensus-first) | Done | `raft.go`: `Propose` |
| PrevLog consistency check (O(1)) | Done | `raft.go`: `HandleAppendEntries` |
| Log truncation (conflict resolution) | Done | `store.go`: `TruncateWALFrom` |
| Commit index advancement | Done | `raft.go`: `advanceCommitIndex` |
| Apply committed entries to tree | Done | `raft.go`: `applyCommitted` |
| Storage interface (decoupled from Store) | Done | `raft.go`: `Storage` interface |
| 24 unit + integration tests | Done | `raft_test.go` |

---

## Known Bugs and Issues

### 1. Disk WAL Truncation Not Implemented

**Severity: Medium**

`TruncateWALFrom()` only truncates the in-memory cache (`entries []wal.Entry`). The disk WAL file still has the old/conflicting entries. New entries appended after truncation will have the same TxIDs as the old ones.

**Why it works for now**: On restart, `Store.replay()` replays ALL WAL entries in order. Later entries with the same TxID overwrite earlier ones in the tree. But this is fragile — a duplicate CREATE would fail during replay (though errors are ignored).

**Fix needed**: Either truncate the actual file, or rewrite the WAL file after truncation. Alternatively, rely on snapshots to capture the correct state and truncate the WAL file after snapshot.

### 2. Cluster-Aware Recovery Not Implemented

**Severity: Medium**

On restart, `Store.replay()` applies ALL WAL entries to the DataTree. In cluster mode, only **committed** entries should be applied. A follower might have uncommitted entries from a deposed leader that should NOT be applied.

**Current behavior**: Uncommitted entries get applied during replay → tree has data that was never committed → inconsistency.

**Fix needed**: Store `commitIndex` durably (in snapshot or separate file). During replay, only apply entries up to the stored commitIndex.

### 3. No gRPC Transport (Still Using fakeTransport)

**Severity: High — blocks deployment**

The cluster only works in tests via `fakeTransport` (direct method calls). There is no real network communication between nodes.

**Fix needed**: Implement a `gRPCTransport` that sends `AppendEntries` and `RequestVote` over gRPC to other nodes. Requires a new proto service for Raft messages (separate from the client-facing ZooKeeper service).

### 4. gRPC Server Bypasses Raft

**Severity: High — blocks deployment**

The gRPC server (`internal/server/server.go`) calls `store.Create()` / `store.Set()` / `store.Delete()` directly. In cluster mode, writes MUST go through `RaftNode.Propose()` for replication.

**Current behavior**: Writes go directly to the local Store, bypassing consensus. Other nodes don't learn about the write. Data diverges.

**Fix needed**: In cluster mode, the server should call `raftNode.Propose(op, path, data)` for writes. Reads can still go to the local Store (or only to the leader for linearizable reads).

### 5. No Log Compaction

**Severity: Low (for now)**

The in-memory WAL cache (`entries []wal.Entry`) grows forever. After a snapshot, entries covered by the snapshot could be discarded.

**Current behavior**: Memory usage grows linearly with the number of writes. Not a problem for small datasets, but will be for long-running production clusters.

**Fix needed**: After `TakeSnapshot()`, truncate entries from the cache that are covered by the snapshot. Update `GetWALEntriesFrom` to handle the case where requested entries have been compacted.

### 6. No InstallSnapshot (Slow Follower Recovery)

**Severity: Low (for now)**

If a follower is so far behind that the leader has already compacted the entries it needs, the leader can't send them. The Raft paper's solution is `InstallSnapshot` — send the full snapshot to the slow follower.

**Current behavior**: The leader returns `nil` from `GetWALEntriesFrom()` and the follower never catches up.

**Fix needed**: Detect when a follower needs entries that have been compacted. Send the leader's snapshot + any entries after it.

### 7. Flaky Automatic Election Test

**Severity: Low**

`TestElection_Automatic` is timing-dependent. It starts 3 nodes with short timeouts, sleeps 500ms, and checks for exactly one leader. On slow machines or under load, this can occasionally fail.

**Current behavior**: Passes reliably in most runs, but has been observed to fail once.

**Fix needed**: Use a deterministic approach (e.g., channels or condition variables) instead of `time.Sleep`. Or increase the sleep margin.

### 8. Raft State Not Persisted

**Severity: Medium**

`CurrentTerm` and `VotedFor` live only in memory. The Raft paper requires these to survive restarts — otherwise a node could vote twice in the same term after a restart, potentially electing two leaders.

**Current behavior**: After restart, node starts at term 0 with no vote. If other nodes are still in a higher term, this works out (the restarted node adopts the higher term). But edge cases exist.

**Fix needed**: Persist `CurrentTerm` and `VotedFor` to disk (in a small file or in the WAL metadata) before responding to any message.

### 9. LastLogTerm Not Used in Vote Comparison

**Severity: Low**

The Raft paper's "up-to-date" check for votes compares both the last log entry's **term** and **index**. Our implementation only compares `LastLogTxID` (index). A candidate with a higher last-term but lower last-index should still win the comparison.

**Current behavior**: Works correctly in most cases because term conflicts are resolved via prevLog. But technically violates the Raft paper's election safety property.

**Fix needed**: Add `LastLogTerm` to `RequestVoteRequest`. Compare term first, then index.

---

## Roadmap to Production

### Phase 2 Remaining: Make the Cluster Actually Run

| Task | Description | Depends On |
|------|-------------|------------|
| **gRPC Transport** | Implement `gRPCTransport` that sends Raft messages over the network. New proto service with `AppendEntries` and `RequestVote` RPCs. | — |
| **Update gRPC server** | In cluster mode, route writes through `RaftNode.Propose()` instead of `Store.Create/Set/Delete`. | gRPC Transport |
| **Update zknode main** | Create `RaftNode` + `Store` + `gRPCTransport`. Accept cluster flags (`--node-id`, `--peers`). | gRPC Transport |
| **Persist Raft state** | Write `CurrentTerm` + `VotedFor` to disk before responding to messages. | — |
| **Cluster-aware recovery** | Store commitIndex durably. On restart, only apply committed entries. | Persist Raft state |
| **Disk WAL truncation** | Truncate the actual WAL file when `TruncateWALFrom` is called. | — |
| **Fix vote comparison** | Add `LastLogTerm` to `RequestVoteRequest` for correct up-to-date check. | — |

### Phase 3: Log Shipping and Compaction

| Task | Description |
|------|-------------|
| **Log compaction** | After snapshot, discard WAL entries covered by the snapshot. |
| **InstallSnapshot** | Send snapshot to followers that are too far behind. |
| **Follower read forwarding** | Followers forward write requests to the leader. |
| **Leader lease / read index** | Linearizable reads without full consensus round. |

### Phase 4: Sessions and Watches

| Task | Description |
|------|-------------|
| **Client sessions** | Persistent connections with session IDs and timeouts. |
| **Heartbeat keepalive** | Clients send periodic heartbeats to keep sessions alive. |
| **Watch notifications** | Clients register watches on znodes, get notified on changes. |
| **Ephemeral nodes** | Znodes that disappear when the session that created them expires. |

### Phase 5: Distributed Primitives

| Task | Description |
|------|-------------|
| **Distributed locks** | Mutual exclusion using sequential znodes. |
| **Leader election primitive** | Using ephemeral + sequential znodes. |
| **Configuration management** | Watch-based config push to all nodes. |
| **Service discovery** | Register/discover services via ephemeral znodes. |

### Phase 6: Deployment

| Task | Description |
|------|-------------|
| **Docker containerization** | Dockerfile for zknode. |
| **Docker Compose** | 3-node cluster with a single `docker-compose up`. |
| **Kubernetes manifests** | StatefulSet for a production-like deployment. |
| **Monitoring** | Metrics (Prometheus), health checks, readiness probes. |
| **Benchmarking** | Throughput and latency measurements. |

---

## Test Summary

| Package | Tests | Description |
|---------|-------|-------------|
| `internal/znode` | 8 | Tree operations: create, get, set, delete, children, errors |
| `internal/wal` | 4 | WAL: append, read all, recovery, TxID sequencing |
| `internal/snapshot` | 3 | Snapshot: save, load, no-file handling |
| `internal/store` | 8 | Store: CRUD, WAL+tree coordination, snapshot recovery, restart |
| `internal/cluster` | 24 | Raft: elections, replication, commit, apply, consistency, integration |
| **Total** | **47** | |

Run all tests:
```
go test ./...
```
