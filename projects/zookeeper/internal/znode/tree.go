package znode

import (
	"fmt"
	"strings"

	"github.com/syamsularifin/zookeeper/internal/snapshot"
)

// DataTree is the container that holds the entire znode tree.
// It has a root node ("/") and provides operations to create/get znodes by path.
//
// Why do we need this wrapper? Why not just use ZNode directly?
//
// Because clients work with PATHS like "/app/config", but the tree is made
// of parent-child POINTERS. Someone needs to translate:
//   path "/app/config"  →  root.Children["app"].Children["config"]
//
// That's DataTree's job: it walks the tree to find the right node.
type DataTree struct {
	// root is the "/" node. It always exists and can never be deleted.
	// Every path starts from here.
	root *ZNode
}

// NewDataTree creates an empty tree with just the root node.
func NewDataTree() *DataTree {
	return &DataTree{
		root: &ZNode{
			// Root starts with an empty Children map, ready for children.
			// make() is important here — a nil map would panic on assignment.
			Children: make(map[string]*ZNode),
		},
	}
}

// Create adds a new znode at the given path.
//
// Rules (same as a real filesystem):
//   - Path must start with "/"
//   - Parent must exist: you can't create "/a/b" if "/a" doesn't exist
//   - Node must not already exist: no overwriting via Create
//
// Example:
//   tree.Create("/app", []byte("hello"))       // OK
//   tree.Create("/app/config", []byte("..."))   // OK — /app exists
//   tree.Create("/x/y/z", []byte("..."))        // ERROR — /x doesn't exist
//   tree.Create("/app", []byte("again"))         // ERROR — /app already exists
func (dt *DataTree) Create(path string, data []byte) error {
	// Step 1: Split the path into parent and child name.
	//
	// "/app/config" → parent="/app", name="config"
	// "/app"        → parent="/",    name="app"
	//
	// We need both because:
	//   - We must find the PARENT node (to add the child to it)
	//   - We need the NAME to use as the map key in parent.Children
	parentPath, name := splitPath(path)

	// Step 2: Walk the tree to find the parent node.
	parent, err := dt.findNode(parentPath)
	if err != nil {
		return fmt.Errorf("parent does not exist: %w", err)
	}

	// Step 3: Check if the node already exists.
	if _, exists := parent.Children[name]; exists {
		return fmt.Errorf("node %q already exists", path)
	}

	// Step 4: Create the new node and attach it to the parent.
	parent.Children[name] = &ZNode{
		Data:     data,
		Children: make(map[string]*ZNode), // ready for its own children
	}

	return nil
}

// Get retrieves the data stored at the given path.
//
// It returns a COPY of the data, not a reference to the original.
// Why? If we returned the original slice, the caller could modify it
// and corrupt our tree without going through proper channels.
func (dt *DataTree) Get(path string) ([]byte, error) {
	node, err := dt.findNode(path)
	if err != nil {
		return nil, err
	}

	// Make a copy so the caller can't modify our internal data.
	dataCopy := make([]byte, len(node.Data))
	copy(dataCopy, node.Data)

	return dataCopy, nil
}

// Set updates the data of an existing znode.
//
// Unlike Create, Set does NOT create a new node — the node must already exist.
// This is intentional: Create and Set are separate operations so you can tell
// the difference between "this is new" vs "this is an update".
func (dt *DataTree) Set(path string, data []byte) error {
	node, err := dt.findNode(path)
	if err != nil {
		return err
	}

	node.Data = data
	return nil
}

// Delete removes a znode from the tree.
//
// Important rule: you can only delete a node that has NO children.
// Why? Safety. If we allowed deleting "/app" while "/app/config" exists,
// the child becomes unreachable (orphaned). ZooKeeper forces you to delete
// children first, so you're always explicit about what you're removing.
//
// To delete a subtree, delete from the leaves up:
//   tree.Delete("/app/config")  // leaf first
//   tree.Delete("/app")         // now safe
func (dt *DataTree) Delete(path string) error {
	if path == "/" {
		return fmt.Errorf("cannot delete root node")
	}

	// We need the PARENT to remove the child from parent.Children.
	// We can't delete a node by just finding it — we need to unlink it
	// from its parent's map.
	parentPath, name := splitPath(path)

	parent, err := dt.findNode(parentPath)
	if err != nil {
		return fmt.Errorf("parent does not exist: %w", err)
	}

	child, exists := parent.Children[name]
	if !exists {
		return fmt.Errorf("node %q does not exist", path)
	}

	// Block deletion if the node has children (no orphans allowed)
	if len(child.Children) > 0 {
		return fmt.Errorf("node %q has children, delete them first", path)
	}

	// Remove from parent's map. After this, nothing references the node
	// and Go's garbage collector will free it.
	delete(parent.Children, name)
	return nil
}

// GetChildren returns the names of all direct children of the node at path.
//
// Example:
//   tree has: /app, /app/config, /app/leader
//   tree.GetChildren("/app") → ["config", "leader"]
//
// This is how service discovery works: create children under /services,
// then GetChildren("/services") returns all registered services.
func (dt *DataTree) GetChildren(path string) ([]string, error) {
	node, err := dt.findNode(path)
	if err != nil {
		return nil, err
	}

	// Collect the map keys into a slice
	children := make([]string, 0, len(node.Children))
	for name := range node.Children {
		children = append(children, name)
	}

	return children, nil
}

// findNode walks the tree from root to find the node at the given path.
//
// How it works for path "/app/config":
//
//   1. Strip the leading "/" → "app/config"
//   2. Split by "/" → ["app", "config"]
//   3. Start at root
//   4. Look up root.Children["app"] → found, move to that node
//   5. Look up node.Children["config"] → found, return it
//
// If any step fails (child doesn't exist), return an error.
func (dt *DataTree) findNode(path string) (*ZNode, error) {
	// Special case: "/" means the root itself.
	if path == "/" {
		return dt.root, nil
	}

	// "/app/config" → "app/config" → ["app", "config"]
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	current := dt.root
	for _, part := range parts {
		child, exists := current.Children[part]
		if !exists {
			return nil, fmt.Errorf("node %q does not exist", path)
		}
		current = child
	}

	return current, nil
}

// splitPath splits a path into its parent path and the last segment (name).
//
// Examples:
//   "/app/config/key" → ("/app/config", "key")
//   "/app"            → ("/", "app")
//
// This is like filepath.Split but for our znode paths.
func splitPath(path string) (string, string) {
	// Remove any trailing slash: "/app/" → "/app"
	path = strings.TrimRight(path, "/")

	// Find the last "/": this separates parent from name
	idx := strings.LastIndex(path, "/")

	// If "/" is at position 0, parent is root
	// "/app" → idx=0 → parent="/", name="app"
	if idx == 0 {
		return "/", path[1:]
	}

	// "/app/config" → idx=4 → parent="/app", name="config"
	return path[:idx], path[idx+1:]
}

// ToSnapshot walks the tree and collects every znode into a flat list.
//
// We walk depth-first, which means parents always appear before children:
//
//   Tree:                    Output order:
//   /                        1. /
//   ├── app "hello"          2. /app
//   │   └── config "5432"    3. /app/config
//   └── locks                4. /locks
//
// Why parents first? Because RestoreFromSnapshot creates nodes in order.
// To create /app/config, /app must already exist. If we listed children
// before parents, the restore would fail with "parent does not exist".
func (dt *DataTree) ToSnapshot() []snapshot.NodeData {
	var nodes []snapshot.NodeData
	dt.walkNode("/", dt.root, &nodes)
	return nodes
}

// walkNode is a recursive helper for ToSnapshot.
// It adds the current node, then recurses into each child.
//
// For the tree:
//   /
//   ├── app
//   │   └── config
//   └── locks
//
// The calls look like:
//   walkNode("/",          root)       → adds "/"
//     walkNode("/app",     app_node)   → adds "/app"
//       walkNode("/app/config", config_node) → adds "/app/config"
//     walkNode("/locks",   locks_node) → adds "/locks"
func (dt *DataTree) walkNode(path string, node *ZNode, nodes *[]snapshot.NodeData) {
	// Add this node to the list
	*nodes = append(*nodes, snapshot.NodeData{
		Path: path,
		Data: node.Data,
	})

	// Recurse into children
	for name, child := range node.Children {
		childPath := path + "/" + name
		// Special case: root is "/" not "", so "/app" not "//app"
		if path == "/" {
			childPath = "/" + name
		}
		dt.walkNode(childPath, child, nodes)
	}
}

// RestoreFromSnapshot rebuilds the tree from a flat list of NodeData.
//
// It walks through the list in order, creating each node.
// Because ToSnapshot outputs parents before children (depth-first),
// we can simply call Create for each entry and the parent will always
// exist by the time we reach the child.
//
// Example:
//   Input: [{"/", nil}, {"/app", "hello"}, {"/app/config", "5432"}]
//
//   Step 1: "/" → skip (root always exists)
//   Step 2: "/app" → Create("/app", "hello")  ✓ (parent "/" exists)
//   Step 3: "/app/config" → Create("/app/config", "5432")  ✓ (parent "/app" exists)
func (dt *DataTree) RestoreFromSnapshot(nodes []snapshot.NodeData) {
	// Reset the tree to empty (just root)
	dt.root = &ZNode{
		Children: make(map[string]*ZNode),
	}

	for _, nd := range nodes {
		if nd.Path == "/" {
			// Root always exists — just restore its data
			dt.root.Data = nd.Data
			continue
		}

		// Create the node. We reuse our existing Create method.
		dt.Create(nd.Path, nd.Data)
	}
}
