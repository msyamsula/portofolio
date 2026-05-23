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

	"github.com/syamsularifin/zookeeper/internal/wal"
	"github.com/syamsularifin/zookeeper/internal/znode"
)

// Store is the durable data store. All mutations go through here.
type Store struct {
	tree *znode.DataTree
	wal  *wal.WAL
}

// New creates a Store with a fresh DataTree and opens the WAL file.
// If the WAL file has existing entries, it replays them to rebuild the tree.
func New(walPath string) (*Store, error) {
	w, err := wal.Open(walPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAL: %w", err)
	}

	s := &Store{
		tree: znode.NewDataTree(),
		wal:  w,
	}

	// Replay existing WAL entries to rebuild the tree.
	// This is what happens on restart:
	//   1. Open the WAL file (entries are still there on disk)
	//   2. Read each entry
	//   3. Apply it to the empty tree
	//   4. Tree is now back to its last state
	if err := s.replay(); err != nil {
		w.Close()
		return nil, fmt.Errorf("WAL replay failed: %w", err)
	}

	return s, nil
}

// replay reads all WAL entries and applies them to the tree.
// Called once during startup to recover state.
func (s *Store) replay() error {
	entries, err := s.wal.ReadAll()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Apply each entry to the tree. We ignore errors here because
		// some operations might "fail" during replay:
		//
		// Example: WAL has CREATE /app, then DELETE /app, then CREATE /app.
		// All 3 succeed on replay. But if the process crashed after the
		// WAL write but before the tree apply, we might have a WAL entry
		// for a CREATE that was never applied, followed by a DELETE.
		// The DELETE would fail ("not found"), and that's fine — we skip it.
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

// Close shuts down the store and closes the WAL file.
func (s *Store) Close() error {
	return s.wal.Close()
}
