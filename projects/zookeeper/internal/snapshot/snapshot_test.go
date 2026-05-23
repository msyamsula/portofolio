package snapshot

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snapshot.json")

	// Create a snapshot
	original := &Snapshot{
		TxID:      42,
		Timestamp: time.Now(),
		Nodes: []NodeData{
			{Path: "/"},
			{Path: "/app", Data: []byte("hello")},
			{Path: "/app/config", Data: []byte("port=5432")},
		},
	}

	// Save it
	err := Save(path, original)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load it back
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify
	if loaded.TxID != 42 {
		t.Fatalf("expected TxID=42, got %d", loaded.TxID)
	}
	if len(loaded.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(loaded.Nodes))
	}
	if string(loaded.Nodes[1].Data) != "hello" {
		t.Fatalf("expected 'hello', got '%s'", string(loaded.Nodes[1].Data))
	}
}

func TestLoadNoFile(t *testing.T) {
	// Loading a file that doesn't exist should return nil, not an error.
	// This is normal on first boot — no snapshot has been taken yet.
	snap, err := Load("/tmp/does-not-exist-snapshot.json")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if snap != nil {
		t.Fatal("expected nil snapshot for missing file")
	}
}
