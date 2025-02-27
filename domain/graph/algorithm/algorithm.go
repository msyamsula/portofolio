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
}

func New() *Service {
	return &Service{}
}
