package znode

// Think of ZooKeeper as a tiny filesystem that lives in memory.
// In a filesystem, you have files and folders.
// In ZooKeeper, you have "znodes" — and every znode is BOTH a file AND a folder.
//
// Example tree:
//
//   /                        <- root (always exists)
//   /app                     <- holds data "my-app-v2", and has children below
//   /app/leader              <- holds data "server-3"
//   /app/config              <- holds data "port=5432"
//
// That's it. A znode = a path + some data + children.

// ZNode is one node in the tree. This is the smallest building block.
type ZNode struct {
	// Data is the actual value stored at this node.
	// It's just raw bytes — could be a string, JSON, anything.
	// Example: []byte("server-3") to mark who the leader is.
	Data []byte

	// Children maps a child name to its ZNode.
	// For example, if this node is "/app", Children might contain:
	//   "leader" -> &ZNode{Data: []byte("server-3")}
	//   "config" -> &ZNode{Data: []byte("port=5432")}
	//
	// We use a map[string]*ZNode, not a slice, because:
	//   - We need fast lookup by name: children["leader"]
	//   - We need fast existence check: _, exists := children["leader"]
	//   - We need fast delete: delete(children, "leader")
	Children map[string]*ZNode
}
