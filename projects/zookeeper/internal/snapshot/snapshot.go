package snapshot

// THE PROBLEM:
//
// The WAL file grows forever. Every Create, Set, Delete adds a line.
// After a month of running, it could have millions of lines.
// On restart, replaying all of them is slow.
//
// Worse: many entries cancel out. If you create and delete /temp
// a thousand times, that's 2000 WAL entries with zero net effect.
//
// THE IDEA:
//
// A snapshot is a photo of the tree at one moment in time.
// Instead of replaying 1 million WAL entries, just load the photo.
//
// Example:
//
//   WAL has 1,000,000 entries (tx_id 1 through 1,000,000)
//   Snapshot taken at tx_id 800,000 (captures the full tree at that moment)
//
//   On restart:
//     Old way: replay entries 1 to 1,000,000  (slow)
//     New way: load snapshot + replay entries 800,001 to 1,000,000  (fast)
//
// That's it. A snapshot = the tree state dumped to a file.

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// NodeData is one znode serialized to a flat format.
//
// The tree in memory is a nested structure (parent → children pointers).
// We flatten it into a list for the snapshot file:
//
//   Tree in memory:           Snapshot file:
//   /                         [
//   ├── app "hello"             {path: "/",          data: nil},
//   │   └── config "port=5432"    {path: "/app",       data: "hello"},
//   └── locks nil               {path: "/app/config", data: "port=5432"},
//                               {path: "/locks",      data: nil},
//                             ]
//
// Why flatten? Because JSON can't easily represent pointer-based trees.
// A flat list is simple to write, simple to read, simple to debug.
type NodeData struct {
	Path string `json:"path"`
	Data []byte `json:"data,omitempty"`
}

// Snapshot is the full snapshot written to disk.
type Snapshot struct {
	// TxID is the WAL transaction ID when this snapshot was taken.
	// On recovery: skip all WAL entries with TxID <= this value.
	// They're already captured in the snapshot.
	TxID int64 `json:"tx_id"`

	// Timestamp is when the snapshot was created. For debugging only.
	Timestamp time.Time `json:"timestamp"`

	// Nodes is every znode in the tree, flattened into a list.
	Nodes []NodeData `json:"nodes"`
}

// Save writes a snapshot to disk.
//
// We use a safe write pattern called "write-to-temp-then-rename":
//
//   1. Write to "snapshot.json.tmp"   (temporary file)
//   2. Sync the temp file to disk     (force it to physical storage)
//   3. Rename tmp → "snapshot.json"   (atomic swap)
//
// Why not just write directly to "snapshot.json"?
//
// If we crash mid-write, the file is half-written and corrupted.
// On next startup, Load() reads a broken file and fails.
// The node can never start again — that's catastrophic.
//
// With the temp file approach:
//   - Crash during step 1 or 2: tmp is corrupted, but snapshot.json
//     still has the PREVIOUS good snapshot. Node starts fine.
//   - Step 3 (rename) is atomic on most filesystems: the file is either
//     fully the old version or fully the new version, never half-and-half.
func Save(path string, snap *Snapshot) error {
	tmpPath := path + ".tmp"

	// Marshal to pretty JSON so we can read it for debugging.
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	// Write to temp file
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Sync: force the temp file to physical disk before renaming.
	// Without this, the OS might rename the file but not yet flush
	// the contents — crash at that moment = empty file with the right name.
	f, err := os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to open temp file for sync: %w", err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	f.Close()

	// Atomic rename: old snapshot is replaced in one instant.
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}

	return nil
}

// Load reads a snapshot from disk.
// Returns nil (not an error) if the file doesn't exist — that's normal
// on first boot when no snapshot has been taken yet.
func Load(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no snapshot yet, that's fine
		}
		return nil, fmt.Errorf("failed to read snapshot: %w", err)
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot: %w", err)
	}

	return &snap, nil
}
