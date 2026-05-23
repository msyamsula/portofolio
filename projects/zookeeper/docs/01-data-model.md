# 01 - Data Model: Znodes

## The Problem

Distributed systems need a shared place to store small pieces of coordination data:

- "server-3 is the current leader"
- "the database is at 10.0.0.5:5432"
- "client A holds the lock on resource X"

This data is tiny (bytes to kilobytes), but every node in the cluster must agree on it.

## The Solution: A Tree of Znodes

ZooKeeper organizes data as a tree, like a filesystem:

```
/
├── app
│   ├── leader         data: "server-3"
│   └── config         data: "port=5432"
├── locks
└── services
    ├── api-server-1   data: "10.0.1.1:8080"
    └── api-server-2   data: "10.0.1.2:8080"
```

Each node in this tree is called a **znode**. A znode is both a file (holds data) and a folder (has children). This is different from a regular filesystem where a folder can't hold data.

## Why a Tree, Not a Flat Key-Value Store?

A flat store like Redis would map keys to values:

```
"/app/leader"  → "server-3"
"/app/config"  → "port=5432"
```

A tree gives us operations that a flat store can't do efficiently:

- **GetChildren("/services")** - list all service instances instantly. In a flat store, you'd need to scan all keys starting with "/services/".
- **Create("/a/b")** - fails if "/a" doesn't exist. This enforces structure. In a flat store, any key can exist independently.
- **Delete("/app")** - fails if it has children. This prevents accidentally orphaning data.

## The ZNode Struct

```go
type ZNode struct {
    Data     []byte
    Children map[string]*ZNode
}
```

Two fields. That's the whole thing.

- `Data` is the value: any bytes, typically a short string.
- `Children` maps child names to child nodes. This creates the tree structure through pointers.

## The DataTree

Clients work with paths like "/app/config". The `DataTree` translates paths into tree traversals:

```
tree.Create("/app/config", []byte("port=5432"))

Internally:
  1. splitPath("/app/config") → parent="/app", name="config"
  2. findNode("/app")         → walks root.Children["app"]
  3. parent.Children["config"] = &ZNode{Data: "port=5432"}
```

### findNode - the tree walker

```
findNode("/app/config")

  "/app/config"
  → TrimPrefix → "app/config"
  → Split("/") → ["app", "config"]
  → start at root
  → root.Children["app"]    → found, move here
  → node.Children["config"] → found, return it
```

### Operations

| Operation | What it does | Rules |
|-----------|-------------|-------|
| Create(path, data) | Add a new znode | Parent must exist. Node must not exist. |
| Get(path) | Read data | Returns a copy (not a reference). |
| Set(path, data) | Update data | Node must already exist. |
| Delete(path) | Remove a znode | Must have no children. Cannot delete root. |
| GetChildren(path) | List child names | Returns names, not full paths. |

### Why Get Returns a Copy

```go
dataCopy := make([]byte, len(node.Data))
copy(dataCopy, node.Data)
return dataCopy, nil
```

Go slices are references to an underlying array. If we returned `node.Data` directly, the caller could modify it and corrupt our tree without going through `Set()`. The copy prevents this.

### Why Delete Needs the Parent

To delete a node, we remove it from `parent.Children`:

```go
delete(parent.Children, name)
```

A node can't remove itself from its parent's map. We must find the parent first, then remove the child entry. That's why Delete splits the path and finds the parent, not the node itself.

## Files

- `internal/znode/znode.go` - ZNode struct
- `internal/znode/tree.go` - DataTree with all operations
- `internal/znode/tree_test.go` - Tests
