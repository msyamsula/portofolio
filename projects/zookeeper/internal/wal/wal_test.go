package wal

import (
	"path/filepath"
	"testing"
)

func TestAppendAndReadAll(t *testing.T) {
	// t.TempDir() creates a temporary directory that Go deletes after the test.
	// Perfect for test files — no cleanup needed.
	dir := t.TempDir()
	path := filepath.Join(dir, "test.wal")

	w, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer w.Close()

	// Append 3 entries
	w.Append(Entry{Op: OpCreate, Path: "/app", Data: []byte("hello")})
	w.Append(Entry{Op: OpSet, Path: "/app", Data: []byte("world")})
	w.Append(Entry{Op: OpDelete, Path: "/app"})

	// Read them back
	entries, err := w.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	// Should have 3 entries
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	// TxIDs should be 1, 2, 3
	for i, e := range entries {
		expected := int64(i + 1)
		if e.TxID != expected {
			t.Fatalf("entry %d: expected TxID=%d, got %d", i, expected, e.TxID)
		}
	}

	// Verify content
	if entries[0].Op != OpCreate {
		t.Fatalf("expected CREATE, got %s", entries[0].Op)
	}
	if string(entries[0].Data) != "hello" {
		t.Fatalf("expected 'hello', got '%s'", string(entries[0].Data))
	}
	if entries[2].Op != OpDelete {
		t.Fatalf("expected DELETE, got %s", entries[2].Op)
	}
}

func TestWALSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.wal")

	// Simulate first run: write 2 entries, then "crash" (close)
	w1, _ := Open(path)
	w1.Append(Entry{Op: OpCreate, Path: "/a", Data: []byte("1")})
	w1.Append(Entry{Op: OpCreate, Path: "/b", Data: []byte("2")})
	w1.Close()

	// Simulate restart: open the same file, read back
	w2, _ := Open(path)
	entries, _ := w2.ReadAll()

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries after restart, got %d", len(entries))
	}

	if string(entries[0].Data) != "1" {
		t.Fatalf("expected '1', got '%s'", string(entries[0].Data))
	}
	if string(entries[1].Data) != "2" {
		t.Fatalf("expected '2', got '%s'", string(entries[1].Data))
	}

	// New appends should continue from TxID=3, not TxID=1
	txID, _ := w2.Append(Entry{Op: OpCreate, Path: "/c"})
	if txID != 3 {
		t.Fatalf("expected TxID=3, got %d", txID)
	}

	w2.Close()
}

func TestOpenExistingWALSafeWithoutReadAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.wal")

	// First run: write 3 entries
	w1, _ := Open(path)
	w1.Append(Entry{Op: OpCreate, Path: "/a"})
	w1.Append(Entry{Op: OpCreate, Path: "/b"})
	w1.Append(Entry{Op: OpCreate, Path: "/c"})
	w1.Close()

	// Second run: open the same file, Append immediately WITHOUT calling ReadAll.
	// Before the fix, this would assign TxID=1 (duplicate).
	// After the fix, Open scans the file, so nextTxID is already 4.
	w2, _ := Open(path)
	txID, _ := w2.Append(Entry{Op: OpCreate, Path: "/d"})

	if txID != 4 {
		t.Fatalf("expected TxID=4, got %d (Open didn't scan existing entries)", txID)
	}

	w2.Close()
}
