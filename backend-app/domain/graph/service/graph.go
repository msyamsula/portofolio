package service

import "strconv"

// Graph represents a graph data structure with nodes and edges
type Graph struct {
	Grabber    map[string]*Node
	IsDirected bool
}

// Node represents a node in the graph
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

// Color represents the color of a node for graph traversal
type Color int

const (
	White Color = iota
	Grey
	Black
)

// CyclePair represents a pair of nodes that form a cycle
type CyclePair struct {
	Start *Node
	End   *Node
}

// GetNode retrieves a node from the graph by ID
func GetNode(g *Graph, id string) *Node {
	return g.Grabber[id]
}

// NewGraph creates a new graph from nodes and edges
func NewGraph(nodes []string, edges [][]string, isDirected ...bool) *Graph {
	directed := false
	if len(isDirected) > 0 {
		directed = isDirected[0]
	}
	g := &Graph{
		Grabber:    make(map[string]*Node),
		IsDirected: directed,
	}

	for _, n := range nodes {
		g.Grabber[n] = &Node{
			Id:        n,
			Neighbors: make(map[*Node]int),
		}
	}

	for _, e := range edges {
		if len(e) < 2 {
			// bad format continue
			continue
		}

		u, v := e[0], e[1]
		nu, nv := GetNode(g, u), GetNode(g, v)
		if nu == nil || nv == nil {
			continue // bad format
		}

		w := 1
		if len(e) >= 3 {
			var err error
			w, err = strconv.Atoi(e[2])
			if err != nil {
				// bad format continue
				w = 1
			}
		}

		nu.Neighbors[nv] = w
		nv.Indegree++
		nu.Outdegree++

		if g.IsDirected {
			continue // only process explicit edge if directed graph
		}

		nv.Neighbors[nu] = w
		nu.Indegree++
		nv.Outdegree++
	}

	return g
}
