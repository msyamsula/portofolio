package wal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// THE PROBLEM:
//
// Our DataTree lives in memory. If the process crashes, everything is gone.
// We need a way to recover the tree after a restart.
//
// THE IDEA:
//
// Before we change anything in memory, write down WHAT we're about to do
// in a file. That file is the Write-Ahead Log (WAL).
//
// "Write-Ahead" = write to the log AHEAD of (before) the actual change.
//
// If we crash, we can replay the log file to rebuild the tree.
//
// Example WAL file contents (one line per operation):
//
//   {"tx_id":1, "op":"CREATE", "path":"/app", "data":"hello"}
//   {"tx_id":2, "op":"SET",    "path":"/app", "data":"world"}
//   {"tx_id":3, "op":"CREATE", "path":"/app/config", "data":"port=5432"}
//   {"tx_id":4, "op":"DELETE", "path":"/app/config"}
//
// To recover: read line 1, apply it. Read line 2, apply it. And so on.
// After replaying all 4 lines, the tree is back to its last state.

// OpType is the kind of operation. There are only 3 things you can do
// to a tree: create a node, update a node, or delete a node.
type OpType string

const (
	OpCreate OpType = "CREATE"
	OpSet    OpType = "SET"
	OpDelete OpType = "DELETE"
)

// Entry is one line in the WAL file. It records a single operation.
//
// An entry must contain EVERYTHING needed to replay the operation.
// If someone hands you an Entry, you should be able to call
// tree.Create / tree.Set / tree.Delete without needing any other info.
type Entry struct {
	// TxID is the transaction ID — a counter that goes up by 1 for each operation.
	// tx_id=1, tx_id=2, tx_id=3, ...
	//
	// Why do we need this?
	//   1. Ordering — we know which operation happened first
	//   2. Recovery — "I've replayed up to tx_id=42, skip everything before that"
	//   3. Later (Phase 2) — followers tell the leader "send me everything after tx_id=100"
	TxID int64 `json:"tx_id"`

	// Op is what happened: CREATE, SET, or DELETE.
	Op OpType `json:"op"`

	// Path is which znode was affected.
	Path string `json:"path"`

	// Data is the value that was written (only for CREATE and SET).
	// DELETE doesn't need data — it just removes the node.
	//
	// omitempty means: if Data is nil, don't write "data":null to the JSON.
	// Keeps the log file cleaner.
	Data []byte `json:"data,omitempty"`
}

// WAL manages the log file. It can do two things:
//   1. Append — write a new entry to the end of the file
//   2. ReadAll — read all entries from the file (for recovery)
type WAL struct {
	// file is the underlying file handle. We keep it open for the lifetime
	// of the WAL so we can keep appending without reopening.
	file *os.File

	// nextTxID is the counter for the next transaction ID to assign.
	// Starts at 1, goes up by 1 for each Append call.
	nextTxID int64
}

// Open opens (or creates) a WAL file at the given path.
//
// os.O_CREATE — create the file if it doesn't exist (first boot)
// os.O_RDWR   — we need to both read (recovery) and write (append)
// os.O_APPEND — every write goes to the END of the file, never the middle
//
// Why O_APPEND? Two reasons:
//   1. We never want to overwrite old entries — they're our history
//   2. The OS guarantees small appends are atomic — if we crash mid-write,
//      old entries are safe, only the last partial entry is damaged
func Open(path string) (*WAL, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAL file: %w", err)
	}

	w := &WAL{
		file:     file,
		nextTxID: 1,
	}

	// Scan the existing file to find the last TxID.
	// This way, Open is always safe — even if you call Append
	// without calling ReadAll first, TxIDs won't collide.
	lastTxID, err := w.findLastTxID()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to scan existing WAL: %w", err)
	}
	w.nextTxID = lastTxID + 1

	return w, nil
}

// findLastTxID scans the file to find the highest TxID.
// Returns 0 if the file is empty (so nextTxID becomes 1).
func (w *WAL) findLastTxID() (int64, error) {
	// Seek to the start to read the whole file
	if _, err := w.file.Seek(0, 0); err != nil {
		return 0, err
	}

	var lastTxID int64
	scanner := bufio.NewScanner(w.file)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry Entry
		if err := json.Unmarshal(line, &entry); err != nil {
			return 0, err
		}
		lastTxID = entry.TxID
	}

	return lastTxID, scanner.Err()
}

// Append writes one entry to the end of the log file.
//
// The flow:
//   1. Assign a TxID to this entry
//   2. Serialize the entry to JSON
//   3. Write the JSON + a newline to the file
//   4. Call file.Sync() to force it to disk
//   5. Return the assigned TxID
//
// Step 4 (Sync) is the most important and most expensive step.
// Without it, the OS might keep the data in a memory buffer.
// If the machine loses power, that buffer is lost.
// Sync forces the data to the physical disk — after Sync returns,
// the entry survives even a power failure.
func (w *WAL) Append(entry Entry) (int64, error) {
	// Assign the next TxID
	entry.TxID = w.nextTxID
	w.nextTxID++

	// Turn the Entry struct into a JSON byte slice.
	// Example: {"tx_id":1,"op":"CREATE","path":"/app","data":"aGVsbG8="}
	data, err := json.Marshal(entry)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal entry: %w", err)
	}

	// Write JSON + newline. The newline is important:
	// it separates entries so ReadAll can read them line by line.
	//
	//   {"tx_id":1,...}\n
	//   {"tx_id":2,...}\n
	//   {"tx_id":3,...}\n
	data = append(data, '\n')

	if _, err := w.file.Write(data); err != nil {
		return 0, fmt.Errorf("failed to write entry: %w", err)
	}

	// Sync = "flush to physical disk NOW, don't buffer"
	// This is slow (~1-5ms) but guarantees durability.
	if err := w.file.Sync(); err != nil {
		return 0, fmt.Errorf("failed to sync: %w", err)
	}

	return entry.TxID, nil
}

// ReadAll reads every entry from the WAL file and returns them in order.
//
// This is called once at startup to replay the log and rebuild the tree.
//
// How it works:
//   1. Seek to the beginning of the file
//   2. Read line by line (each line = one JSON entry)
//   3. Parse each line into an Entry struct
//   4. Collect them all into a slice
//
// After ReadAll, we also update nextTxID so that new Append calls
// continue from where we left off (not restart at 1).
func (w *WAL) ReadAll() ([]Entry, error) {
	// Seek to the beginning — we might have been appending at the end.
	if _, err := w.file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek: %w", err)
	}

	var entries []Entry
	scanner := bufio.NewScanner(w.file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue // skip empty lines
		}

		var entry Entry
		if err := json.Unmarshal(line, &entry); err != nil {
			return nil, fmt.Errorf("failed to parse entry: %w", err)
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read WAL: %w", err)
	}

	// Update the counter so new appends continue from where we left off.
	// If the file had entries [1, 2, 3], next should be 4.
	if len(entries) > 0 {
		w.nextTxID = entries[len(entries)-1].TxID + 1
	}

	return entries, nil
}

// LastTxID returns the most recently assigned TxID.
// Returns 0 if no entries have been written.
//
// Used by snapshot: "this snapshot covers everything up to TxID X"
func (w *WAL) LastTxID() int64 {
	return w.nextTxID - 1
}

// Close flushes and closes the file.
func (w *WAL) Close() error {
	return w.file.Close()
}
