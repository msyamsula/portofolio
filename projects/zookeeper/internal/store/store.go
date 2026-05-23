package store

// Store ties together DataTree and WAL into a single safe API.
//
// WHY NOT PUT THIS LOGIC IN DataTree OR WAL?
//
// Because they have different jobs:
//   - DataTree knows how to manage znodes in memory (fast reads/writes)
//   - WAL knows how to write entries to a file (durability)
//   - Store knows the ORDER: WAL first, then tree
//
// This separation is called "separation of concerns":
//   DataTree doesn't know about files.
//   WAL doesn't know about znodes.
//   Store knows about both and coordinates them.
//
// THE RULE:
//
// Every write goes through this flow:
//
//   Client calls Store.Create("/app", "hello")
//       ↓
//   Step 1: Write to WAL           ← if this fails, return error (nothing changed)
//       ↓
//   Step 2: Apply to DataTree      ← if this fails, WAL has an extra entry (harmless)
//       ↓
//   Return success to client
//
// Why is an extra WAL entry harmless?
// On replay, the operation would fail with "already exists" or "not found",
// and we just skip it. The WAL is a log of ATTEMPTS, not guarantees.

import (
	"fmt"
	"time"

	"github.com/syamsularifin/zookeeper/internal/snapshot"
	"github.com/syamsularifin/zookeeper/internal/wal"
	"github.com/syamsularifin/zookeeper/internal/znode"
)

// Store is the durable data store. All mutations go through here.
type Store struct {
	tree *znode.DataTree
	wal  *wal.WAL

	// snapPath is where we save/load the snapshot file.
	snapPath string
}

// New creates a Store, recovers from snapshot + WAL, and is ready to serve.
//
// The full recovery flow:
//
//   1. Load snapshot (if exists)  → tree has state at TxID X
//   2. Open WAL                   → read all entries
//   3. Replay WAL entries > X     → tree is fully up to date
//
// On first boot (no snapshot, no WAL), we just start with an empty tree.
func New(walPath string, snapPath string) (*Store, error) {
	s := &Store{
		tree:     znode.NewDataTree(),
		snapPath: snapPath,
	}

	// Step 1: Load snapshot if it exists.
	//
	// snap = nil means no snapshot file (first boot). That's fine.
	// snap != nil means we have a previous state to restore.
	snap, err := snapshot.Load(snapPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load snapshot: %w", err)
	}

	var snapshotTxID int64
	if snap != nil {
		s.tree.RestoreFromSnapshot(snap.Nodes)
		snapshotTxID = snap.TxID
	}

	// Step 2: Open the WAL file.
	w, err := wal.Open(walPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAL: %w", err)
	}
	s.wal = w

	// Step 3: Replay WAL entries that came AFTER the snapshot.
	//
	// If snapshot was at TxID=100, we skip entries 1-100 (already in snapshot)
	// and only replay 101, 102, 103...
	//
	// If there's no snapshot (snapshotTxID=0), we replay everything.
	if err := s.replay(snapshotTxID); err != nil {
		w.Close()
		return nil, fmt.Errorf("WAL replay failed: %w", err)
	}

	return s, nil
}

// replay reads WAL entries and applies those that came after the snapshot.
//
// afterTxID = 0 means "replay everything" (no snapshot).
// afterTxID = 100 means "skip entries 1-100, replay 101+".
func (s *Store) replay(afterTxID int64) error {
	entries, err := s.wal.ReadAll()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Skip entries already covered by the snapshot.
		if entry.TxID <= afterTxID {
			continue
		}

		// Ignore errors — see explanation in applyToTree.
		_ = s.applyToTree(entry)
	}

	return nil
}

// applyToTree applies a single WAL entry to the in-memory tree.
func (s *Store) applyToTree(entry wal.Entry) error {
	switch entry.Op {
	case wal.OpCreate:
		return s.tree.Create(entry.Path, entry.Data)
	case wal.OpSet:
		return s.tree.Set(entry.Path, entry.Data)
	case wal.OpDelete:
		return s.tree.Delete(entry.Path)
	default:
		return fmt.Errorf("unknown operation: %s", entry.Op)
	}
}

// Create adds a new znode. WAL first, then tree.
func (s *Store) Create(path string, data []byte) error {
	// Step 1: WAL — record the intent to disk
	_, err := s.wal.Append(wal.Entry{
		Op:   wal.OpCreate,
		Path: path,
		Data: data,
	})
	if err != nil {
		return fmt.Errorf("WAL write failed: %w", err)
	}

	// Step 2: Tree — apply in memory
	return s.tree.Create(path, data)
}

// Get reads a znode. No WAL needed — reads don't change anything.
func (s *Store) Get(path string) ([]byte, error) {
	return s.tree.Get(path)
}

// Set updates a znode. WAL first, then tree.
func (s *Store) Set(path string, data []byte) error {
	_, err := s.wal.Append(wal.Entry{
		Op:   wal.OpSet,
		Path: path,
		Data: data,
	})
	if err != nil {
		return fmt.Errorf("WAL write failed: %w", err)
	}

	return s.tree.Set(path, data)
}

// Delete removes a znode. WAL first, then tree.
func (s *Store) Delete(path string) error {
	_, err := s.wal.Append(wal.Entry{
		Op:   wal.OpDelete,
		Path: path,
	})
	if err != nil {
		return fmt.Errorf("WAL write failed: %w", err)
	}

	return s.tree.Delete(path)
}

// GetChildren lists children of a znode. No WAL needed — read only.
func (s *Store) GetChildren(path string) ([]string, error) {
	return s.tree.GetChildren(path)
}

// TakeSnapshot saves the current tree state to disk.
//
// When to call this:
//   - Periodically (e.g. every 1000 writes)
//   - Before shutdown (to minimize replay on next startup)
//
// What it does:
//   1. Ask the tree for a flat list of all znodes
//   2. Ask the WAL for the current TxID
//   3. Save both to the snapshot file
//
// After this, on next restart:
//   - Load this snapshot → tree is at TxID X
//   - Replay only WAL entries after X → much faster
func (s *Store) TakeSnapshot() error {
	snap := &snapshot.Snapshot{
		TxID:      s.wal.LastTxID(),
		Timestamp: time.Now(),
		Nodes:     s.tree.ToSnapshot(),
	}

	return snapshot.Save(s.snapPath, snap)
}

// Close takes a final snapshot, then closes the WAL file.
// The snapshot minimizes WAL replay on next startup.
func (s *Store) Close() error {
	s.TakeSnapshot()
	return s.wal.Close()
}
