package service

// Service defines the interface for graph algorithm operations
type Service interface {
	ArticulationPointAndBridge(g *Graph) (log []string, id []string, bridge [][]string)
	BreadthFirstSearch(g *Graph) []string
	DepthFirstSearch(g *Graph) []string
	DirectedAcyclicGraph(g *Graph) (path []string, acyclic bool)
	Eulerian(g *Graph) (path []string)
	IsCycle(g *Graph) (log []string, cycles [][]string)
	StronglyConnectedComponents(g *Graph) (log []string, comp [][]string)
}

type service struct {
	// Algorithm state tracking
	dfsLog  []string
	bfsLog  []string
	cycleLog []string
	cycle    []CyclePair
	dagPath  []string
	sccTree  []string
	apLog    []string
	apId     []string
	bridge   [][]string
	eulerPath []string
	dfsTree  []string
}

// New creates a new graph service
func New() Service {
	return &service{}
}
