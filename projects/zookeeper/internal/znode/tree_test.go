package znode

import "testing"

func TestCreateAndGet(t *testing.T) {
	tree := NewDataTree()

	// Create a node and read it back
	err := tree.Create("/app", []byte("hello"))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	data, err := tree.Get("/app")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(data) != "hello" {
		t.Fatalf("expected 'hello', got '%s'", string(data))
	}
}

func TestCreateNested(t *testing.T) {
	tree := NewDataTree()

	// Create parent first, then child
	tree.Create("/app", []byte("parent"))
	err := tree.Create("/app/config", []byte("port=5432"))
	if err != nil {
		t.Fatalf("Create nested failed: %v", err)
	}

	data, _ := tree.Get("/app/config")
	if string(data) != "port=5432" {
		t.Fatalf("expected 'port=5432', got '%s'", string(data))
	}
}

func TestCreateFailsWithoutParent(t *testing.T) {
	tree := NewDataTree()

	// /a doesn't exist, so /a/b should fail
	err := tree.Create("/a/b", []byte("data"))
	if err == nil {
		t.Fatal("expected error when parent doesn't exist")
	}
}

func TestCreateFailsDuplicate(t *testing.T) {
	tree := NewDataTree()

	tree.Create("/app", nil)
	err := tree.Create("/app", nil)
	if err == nil {
		t.Fatal("expected error on duplicate create")
	}
}

func TestGetFailsNotFound(t *testing.T) {
	tree := NewDataTree()

	_, err := tree.Get("/missing")
	if err == nil {
		t.Fatal("expected error for non-existent node")
	}
}

func TestGetReturnsCopy(t *testing.T) {
	tree := NewDataTree()
	tree.Create("/app", []byte("original"))

	// Get returns a copy — modifying it should NOT affect the tree
	data, _ := tree.Get("/app")
	data[0] = 'X' // mutate the returned slice

	// The tree's data should still be "original"
	data2, _ := tree.Get("/app")
	if string(data2) != "original" {
		t.Fatalf("tree data was corrupted: got '%s'", string(data2))
	}
}

// --- Set tests ---

func TestSet(t *testing.T) {
	tree := NewDataTree()
	tree.Create("/app", []byte("v1"))

	// Update the data
	err := tree.Set("/app", []byte("v2"))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify it changed
	data, _ := tree.Get("/app")
	if string(data) != "v2" {
		t.Fatalf("expected 'v2', got '%s'", string(data))
	}
}

func TestSetFailsNotFound(t *testing.T) {
	tree := NewDataTree()

	err := tree.Set("/missing", []byte("data"))
	if err == nil {
		t.Fatal("expected error when setting non-existent node")
	}
}

// --- Delete tests ---

func TestDelete(t *testing.T) {
	tree := NewDataTree()
	tree.Create("/app", nil)

	err := tree.Delete("/app")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Should be gone now
	_, err = tree.Get("/app")
	if err == nil {
		t.Fatal("node should not exist after delete")
	}
}

func TestDeleteFailsWithChildren(t *testing.T) {
	tree := NewDataTree()
	tree.Create("/app", nil)
	tree.Create("/app/config", nil)

	// Can't delete /app because /app/config still exists
	err := tree.Delete("/app")
	if err == nil {
		t.Fatal("expected error when deleting node with children")
	}
}

func TestDeleteCannotDeleteRoot(t *testing.T) {
	tree := NewDataTree()

	err := tree.Delete("/")
	if err == nil {
		t.Fatal("expected error when deleting root")
	}
}

// --- GetChildren tests ---

func TestGetChildren(t *testing.T) {
	tree := NewDataTree()
	tree.Create("/app", nil)
	tree.Create("/app/config", nil)
	tree.Create("/app/leader", nil)
	tree.Create("/app/locks", nil)

	children, err := tree.GetChildren("/app")
	if err != nil {
		t.Fatalf("GetChildren failed: %v", err)
	}

	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}
}

func TestGetChildrenEmpty(t *testing.T) {
	tree := NewDataTree()
	tree.Create("/app", nil)

	children, err := tree.GetChildren("/app")
	if err != nil {
		t.Fatalf("GetChildren failed: %v", err)
	}

	if len(children) != 0 {
		t.Fatalf("expected 0 children, got %d", len(children))
	}
}

// --- Snapshot tests ---

func TestToSnapshotAndRestore(t *testing.T) {
	// Build a tree
	original := NewDataTree()
	original.Create("/app", []byte("hello"))
	original.Create("/app/config", []byte("port=5432"))
	original.Create("/locks", nil)

	// Dump it to a snapshot
	nodes := original.ToSnapshot()

	// Restore into a NEW empty tree
	restored := NewDataTree()
	restored.RestoreFromSnapshot(nodes)

	// Verify every node survived the round-trip
	data, err := restored.Get("/app")
	if err != nil {
		t.Fatalf("Get /app failed: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected 'hello', got '%s'", string(data))
	}

	data, err = restored.Get("/app/config")
	if err != nil {
		t.Fatalf("Get /app/config failed: %v", err)
	}
	if string(data) != "port=5432" {
		t.Fatalf("expected 'port=5432', got '%s'", string(data))
	}

	_, err = restored.Get("/locks")
	if err != nil {
		t.Fatalf("Get /locks failed: %v", err)
	}
}
