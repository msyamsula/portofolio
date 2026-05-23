package store

import (
	"path/filepath"
	"testing"
)

// helper to create WAL and snapshot paths in the same temp dir
func newTestStore(t *testing.T, dir string) *Store {
	s, err := New(
		filepath.Join(dir, "test.wal"),
		filepath.Join(dir, "snapshot.json"),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	return s
}

func TestCreateAndGet(t *testing.T) {
	s := newTestStore(t, t.TempDir())
	defer s.Close()

	s.Create("/app", []byte("hello"))

	data, err := s.Get("/app")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected 'hello', got '%s'", string(data))
	}
}

func TestDataSurvivesRestart(t *testing.T) {
	dir := t.TempDir()

	// First run: create some data, then shutdown
	s1 := newTestStore(t, dir)
	s1.Create("/app", []byte("hello"))
	s1.Create("/app/config", []byte("port=5432"))
	s1.Set("/app", []byte("updated"))
	s1.Close() // takes snapshot + closes WAL

	// Second run: opens same files — data should be back
	s2 := newTestStore(t, dir)
	defer s2.Close()

	data, err := s2.Get("/app")
	if err != nil {
		t.Fatalf("Get /app failed after restart: %v", err)
	}
	if string(data) != "updated" {
		t.Fatalf("expected 'updated', got '%s'", string(data))
	}

	data, err = s2.Get("/app/config")
	if err != nil {
		t.Fatalf("Get /app/config failed after restart: %v", err)
	}
	if string(data) != "port=5432" {
		t.Fatalf("expected 'port=5432', got '%s'", string(data))
	}
}

func TestDeleteSurvivesRestart(t *testing.T) {
	dir := t.TempDir()

	// First run: create then delete
	s1 := newTestStore(t, dir)
	s1.Create("/temp", []byte("gone soon"))
	s1.Delete("/temp")
	s1.Close()

	// Second run: /temp should NOT exist
	s2 := newTestStore(t, dir)
	defer s2.Close()

	_, err := s2.Get("/temp")
	if err == nil {
		t.Fatal("/temp should not exist after replay — it was deleted")
	}
}

func TestSnapshotSpeedsUpRecovery(t *testing.T) {
	dir := t.TempDir()

	// First run: create data and take a snapshot
	s1 := newTestStore(t, dir)
	s1.Create("/app", []byte("v1"))
	s1.Create("/app/config", []byte("port=5432"))
	s1.TakeSnapshot() // snapshot covers TxID 1 and 2

	// More writes AFTER the snapshot
	s1.Set("/app", []byte("v2")) // TxID 3
	s1.Close()

	// Second run: loads snapshot (TxID 1-2) + replays WAL (TxID 3)
	s2 := newTestStore(t, dir)
	defer s2.Close()

	// /app should be "v2" — the post-snapshot write was replayed
	data, _ := s2.Get("/app")
	if string(data) != "v2" {
		t.Fatalf("expected 'v2', got '%s'", string(data))
	}

	// /app/config should still be there from the snapshot
	data, _ = s2.Get("/app/config")
	if string(data) != "port=5432" {
		t.Fatalf("expected 'port=5432', got '%s'", string(data))
	}
}
