# 04 - Store: The Coordinator

## The Problem

We have three separate components:

- **DataTree** - manages znodes in memory (fast)
- **WAL** - writes operations to disk (durable)
- **Snapshot** - dumps/loads the full tree (fast recovery)

Someone needs to coordinate them. That's the Store.

## The Rule

Every write follows this order:

```
Store.Create("/app", "hello")
    │
    ├── 1. WAL.Append(CREATE /app "hello")   ← disk first
    │
    └── 2. tree.Create("/app", "hello")      ← memory second
```

WAL first, tree second. If we crash between step 1 and 2, the WAL has a record that we can replay on restart. If we did it the other way (tree first, WAL second), a crash would lose data that the client thinks was saved.

Reads skip the WAL entirely:

```
Store.Get("/app")
    │
    └── tree.Get("/app")    ← memory only, fast
```

Reads don't change anything, so there's nothing to log.

## Recovery Flow

On startup, Store recovers the tree from disk:

```
New(walPath, snapPath)

  Step 1: Load snapshot (if exists)
          ├── no file?   → empty tree, snapshotTxID = 0
          └── file found → restore tree, snapshotTxID = 42

  Step 2: Open WAL, read all entries

  Step 3: Replay WAL entries where TxID > snapshotTxID
          snapshotTxID = 0:   replay everything
          snapshotTxID = 42:  skip 1-42, replay 43+

  Ready to serve.
```

### Example

```
First run:
  Create "/app" "v1"          → WAL: tx1
  Create "/app/config" "..."  → WAL: tx2
  TakeSnapshot()              → snapshot.json at tx2
  Set "/app" "v2"             → WAL: tx3
  Close()

Files on disk:
  snapshot.json  → tree state at tx2 (/app="v1", /app/config="...")
  wal.log        → [tx1, tx2, tx3]

Second run (restart):
  Step 1: load snapshot → tree has /app="v1", /app/config="..."
  Step 2: read WAL      → [tx1, tx2, tx3]
  Step 3: skip tx1, tx2 (covered by snapshot), replay tx3 (SET /app "v2")
  Result: /app="v2", /app/config="..."
```

## Why Replay Ignores Errors

During replay, some operations may fail:

```
WAL:
  tx1: CREATE /app     ← succeeds on replay
  tx2: CREATE /app     ← fails: "already exists" (junk entry from a retry)
```

This can happen if the WAL was written but the tree operation failed (e.g. duplicate create). The WAL records attempts, not guarantees. Ignoring replay errors lets the node always start, even with junk entries.

## TakeSnapshot

```go
func (s *Store) TakeSnapshot() error {
    snap := &snapshot.Snapshot{
        TxID:      s.wal.LastTxID(),
        Timestamp: time.Now(),
        Nodes:     s.tree.ToSnapshot(),
    }
    return snapshot.Save(s.snapPath, snap)
}
```

Three things combined:
1. Ask the WAL: what's the latest TxID?
2. Ask the tree: give me all your znodes as a flat list.
3. Save both to disk using the safe write-to-temp-then-rename pattern.

## Close

```go
func (s *Store) Close() error {
    s.TakeSnapshot()      // minimize WAL replay on next startup
    return s.wal.Close()
}
```

Takes a final snapshot before shutting down. This means the next startup replays only the entries written between the snapshot and the shutdown (often zero).

## Separation of Concerns

```
DataTree  → knows about znodes and tree operations
WAL       → knows about files and entries
Snapshot  → knows about serialization and safe disk writes
Store     → knows about all three and the ORDER they run in
```

Each component can be tested independently. The Store is the only place where "WAL before tree" is enforced.

## Files

- `internal/store/store.go` - Store (New, Create, Get, Set, Delete, TakeSnapshot, Close)
- `internal/store/store_test.go` - Tests including restart and snapshot recovery
