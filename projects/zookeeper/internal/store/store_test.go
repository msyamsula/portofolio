package store

import (
	"path/filepath"
	"testing"
)

func TestCreateAndGet(t *testing.T) {
	dir := t.TempDir()
	s, err := New(filepath.Join(dir, "test.wal"))
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
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
	walPath := filepath.Join(dir, "test.wal")

	// First run: create some data, then "crash"
	s1, _ := New(walPath)
	s1.Create("/app", []byte("hello"))
	s1.Create("/app/config", []byte("port=5432"))
	s1.Set("/app", []byte("updated"))
	s1.Close() // simulate shutdown

	// Second run: open the same WAL file — data should be back
	s2, _ := New(walPath)
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
	walPath := filepath.Join(dir, "test.wal")

	// First run: create then delete
	s1, _ := New(walPath)
	s1.Create("/temp", []byte("gone soon"))
	s1.Delete("/temp")
	s1.Close()

	// Second run: /temp should NOT exist
	s2, _ := New(walPath)
	defer s2.Close()

	_, err := s2.Get("/temp")
	if err == nil {
		t.Fatal("/temp should not exist after replay — it was deleted")
	}
}
