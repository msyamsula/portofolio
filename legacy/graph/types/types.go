package types

type GraphNotation struct {
	Nodes []string `json:"nodes,omitempty"`
	Edges []Edge   `json:"edges,omitempty"`
}

type Edge struct {
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	Weight string `json:"weight,omitempty"`
}

type AlgorithmResult struct {
	Log     []string   `json:"log"`
	Path    []string   `json:"path"`
	Cycles  [][]string `json:"cycles"`
	Acyclic bool       `json:"acyclic"`
	Scc     [][]string `json:"scc"`
	Ap      []string   `json:"ap"`
	Bridge  [][]string `json:"bridge"`
}

type Graph struct {
	Grabber    map[string]*Node
	IsDirected bool
}

type Node struct {
	Id        string
	Neighbors map[*Node]int

	// properties
	Visited             bool
	Color               Color
	Parent              *Node
	Indegree, Outdegree int
	In, Out             int // running indegree
	Tin, Tout           int
	Low                 int // lowest reachable node (tin)
	IsArticulationPoint bool
}

type Color int

const (
	White Color = iota
	Grey
	Black
)

type CyclePair struct {
	Start *Node
	End   *Node
}

type Algorithm struct {
	// dfs
	DfsLog []string

	// bfs
	BfsLog []string

	// cycle
	CycleLog   []string
	Cycle      []CyclePair
	CycleExist bool

	// dag
	DagPath []string

	// scc
	SccTree []string

	// ap & bridge
	ApLog  []string
	ApId   []string
	Bridge [][]string

	// eulerian path
	EulerPath []string
	DfsTree   []string
}
