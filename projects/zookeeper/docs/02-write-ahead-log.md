# 02 - Write-Ahead Log (WAL)

## The Problem

The DataTree lives in memory. If the process crashes, everything is gone. We need a way to recover the tree after a restart.

## The Idea

Before changing anything in memory, write down what you're about to do in a file. That file is the Write-Ahead Log.

"Write-Ahead" = write to the log AHEAD of (before) the actual change.

## What the WAL File Looks Like

One JSON line per operation:

```
{"tx_id":1,"op":"CREATE","path":"/app","data":"aGVsbG8="}
{"tx_id":2,"op":"SET","path":"/app","data":"d29ybGQ="}
{"tx_id":3,"op":"CREATE","path":"/app/config","data":"cG9ydD01NDMy"}
{"tx_id":4,"op":"DELETE","path":"/app/config"}
```

Each line is a complete record of one operation. To recover: read line 1, apply it to an empty tree. Read line 2, apply it. Keep going. After replaying all lines, the tree is back to its last state.

## The Entry Struct

```go
type Entry struct {
    TxID int64  `json:"tx_id"`
    Op   OpType `json:"op"`
    Path string `json:"path"`
    Data []byte `json:"data,omitempty"`
}
```

An entry captures everything needed to replay one operation:

- **TxID** - transaction ID, a counter: 1, 2, 3... Never goes backwards.
- **Op** - what happened: CREATE, SET, or DELETE.
- **Path** - which znode was affected.
- **Data** - the value (only for CREATE and SET, DELETE doesn't need it).

### Why int64 for TxID?

int64 max = 9.2 quintillion. At 1 million writes per second, it takes 292,000 years to overflow. Not a concern.

### Why `omitempty` on Data?

DELETE operations have no data. Without omitempty, the JSON would include `"data":null` — unnecessary noise. With it, the field is simply absent, keeping the log clean.

## The WAL Struct

```go
type WAL struct {
    file     *os.File
    nextTxID int64
}
```

Two fields:

- `file` - the open file handle. Stays open for the lifetime of the process.
- `nextTxID` - counter for the next TxID to assign.

## Open - Safe Initialization

```go
func Open(path string) (*WAL, error)
```

Opens (or creates) the WAL file with three flags:

- `O_CREATE` - create the file if it doesn't exist (first boot)
- `O_RDWR` - we need to both read (recovery) and write (append)
- `O_APPEND` - every write goes to the END of the file

After opening, it **scans the file** to find the last TxID. This ensures `nextTxID` is always correct, even if you call `Append` without calling `ReadAll` first.

## Append - The Write Path

```
Append(Entry{Op: OpCreate, Path: "/app", Data: []byte("hello")})

  1. Assign TxID = nextTxID (e.g. 1)
  2. nextTxID++ (now 2)
  3. json.Marshal → {"tx_id":1,"op":"CREATE","path":"/app","data":"aGVsbG8="}
  4. Write JSON + newline to file
  5. file.Sync() ← forces data to physical disk
  6. Return TxID 1
```

### Why file.Sync()?

Without Sync, the OS keeps your data in a RAM buffer. It will eventually write it to disk, but "eventually" could be 30 seconds later. If the machine loses power during that window, the data is gone — even though `Write()` returned success.

`Sync()` forces the OS to flush to the physical disk right now. After Sync returns, the data survives even a power failure.

The trade-off: Sync is slow (~1-5ms). Real systems batch multiple operations before syncing to amortize this cost. We keep it simple: one Sync per Append.

### Why Append-Only?

1. **Fast** - appending is sequential I/O, which is 10-100x faster than random writes.
2. **Safe** - if we crash mid-append, only the last partial entry is damaged. All previous entries are intact.
3. **Simple** - no free space management, no fragmentation, no block allocation.

## ReadAll - The Recovery Path

```
ReadAll()

  1. Seek to beginning of file
  2. Read line by line
  3. Parse each line as JSON → Entry
  4. Update nextTxID = last entry's TxID + 1
  5. Return all entries
```

Called once at startup. After this, the caller can replay each entry against an empty tree to rebuild state.

## Files

- `internal/wal/wal.go` - Entry struct, WAL (Open, Append, ReadAll, Close)
- `internal/wal/wal_test.go` - Tests including crash/restart simulation
