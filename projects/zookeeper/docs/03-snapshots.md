# 03 - Snapshots

## The Problem

The WAL grows forever. Every Create, Set, Delete adds a line. After a month, it could have millions of lines. On restart, replaying all of them is slow.

Worse: many entries cancel out. Creating and deleting `/temp` a thousand times produces 2000 WAL entries with zero net effect.

## The Idea

A snapshot is a photo of the tree at one moment in time. Instead of replaying 1 million WAL entries, just load the photo.

```
WAL has 1,000,000 entries (tx_id 1 through 1,000,000)
Snapshot taken at tx_id 800,000

On restart:
  Without snapshot: replay entries 1 to 1,000,000        (slow)
  With snapshot:    load snapshot + replay 800,001+       (fast)
```

## What the Snapshot File Looks Like

```json
{
  "tx_id": 42,
  "timestamp": "2026-05-23T12:00:00Z",
  "nodes": [
    {
      "path": "/"
    },
    {
      "path": "/app",
      "data": "aGVsbG8="
    },
    {
      "path": "/app/config",
      "data": "cG9ydD01NDMy"
    },
    {
      "path": "/locks"
    }
  ]
}
```

Each field:

- **tx_id: 42** - this snapshot captures the tree state after WAL entry 42. On recovery, skip entries 1-42, replay 43+.
- **timestamp** - when the snapshot was taken. For human debugging only, not used by code.
- **nodes** - flat list of every znode. The `data` values are base64-encoded (Go's JSON encoding for `[]byte`).

## The Structs

```go
type NodeData struct {
    Path string `json:"path"`
    Data []byte `json:"data,omitempty"`
}

type Snapshot struct {
    TxID      int64      `json:"tx_id"`
    Timestamp time.Time  `json:"timestamp"`
    Nodes     []NodeData `json:"nodes"`
}
```

`NodeData` is one znode flattened: just a path and its data. No children pointers, no nesting.

## Save - Safe Write to Disk

```
Save("snapshot.json", snap)

  1. Write to "snapshot.json.tmp"     ← temporary file
  2. Sync tmp to disk                 ← force bytes to physical storage
  3. Rename tmp → "snapshot.json"     ← atomic swap
```

Why not write directly to "snapshot.json"?

If we crash mid-write, the file is half-written and corrupted. On next startup, Load reads a broken file and fails. The node can never start again.

With the temp file approach:
- Crash during step 1 or 2: tmp is corrupted, but snapshot.json still has the previous good snapshot.
- Step 3 (rename) is atomic on POSIX filesystems: the file is either fully the old version or fully the new version, never half-and-half.

## Load - Read from Disk

```
Load("snapshot.json")

  File doesn't exist? → return nil  (first boot, no snapshot yet)
  File exists?        → read → parse JSON → return Snapshot
```

Returning nil (not an error) for a missing file is intentional. On first boot there's no snapshot, and that's normal.

## Tree Serialization

The DataTree has two methods for snapshot integration:

### ToSnapshot - Tree to Flat List

Walks the tree depth-first, collecting every znode:

```
Tree:                         Output:
/                             [{path: "/"}
├── app "hello"                {path: "/app",        data: "hello"}
│   └── config "5432"          {path: "/app/config",  data: "5432"}
└── locks                      {path: "/locks"}]
```

Parents always appear before children. This ordering is critical for RestoreFromSnapshot.

### RestoreFromSnapshot - Flat List to Tree

Walks through the list in order, creating each node:

```
Step 1: "/"           → root always exists, set its data
Step 2: "/app"        → Create("/app", "hello")       works because "/" exists
Step 3: "/app/config" → Create("/app/config", "5432")  works because "/app" exists
Step 4: "/locks"      → Create("/locks", nil)          works because "/" exists
```

Because ToSnapshot outputs parents before children, the parent always exists by the time we create the child.

## Files

- `internal/snapshot/snapshot.go` - NodeData, Snapshot, Save, Load
- `internal/snapshot/snapshot_test.go` - Tests
- `internal/znode/tree.go` - ToSnapshot, RestoreFromSnapshot, walkNode
