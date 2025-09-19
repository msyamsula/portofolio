package service

type CyclePair struct {
	start *Node
	end   *Node
}
type Algorithm struct {
	// dfs
	dfsLog []string

	// bfs
	bfsLog []string

	// cycle
	cycleLog   []string
	cycle      []CyclePair
	cycleExist bool

	// dag
	dagPath []string

	// scc
	sccTree []string

	// ap & bridge
	apLog  []string
	apId   []string
	bridge [][]string

	// eulerian path
	eulerPath []string
	dfsTree   []string
}

func New() *Algorithm {
	return &Algorithm{}
}
