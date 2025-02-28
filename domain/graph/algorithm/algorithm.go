package algorithm

import "github.com/msyamsula/portofolio/domain/graph"

type CyclePair struct {
	start *graph.Node
	end   *graph.Node
}
type Service struct {
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
}

func New() *Service {
	return &Service{}
}
