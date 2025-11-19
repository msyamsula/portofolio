package service

import "github.com/msyamsula/portofolio/backend-app/graph/types"

type Service interface {
	ArticulationPointAndBridge(g *types.Graph) (log []string, id []string, bridge [][]string)
	BreadthFirstSearch(g *types.Graph) []string
	DepthFirstSearch(g *types.Graph) []string
	DirectedAcyclicGraph(g *types.Graph) (path []string, acyclic bool)
	Eulerian(g *types.Graph) (path []string)
	IsCycle(g *types.Graph) (log []string, cycles [][]string)
	StronglyConnectedComponents(g *types.Graph) (log []string, comp [][]string)
}

func New() Service {
	return &service{}
}
