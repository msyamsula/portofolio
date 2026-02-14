package service

import (
	"strconv"

	"github.com/msyamsula/portofolio/backend-app/graph/types"
)

func GetNode(s *types.Graph, id string) *types.Node {
	return s.Grabber[id]
}

func NewGraph(nodes []string, edges [][]string, isDirected ...bool) *types.Graph {
	directed := false
	if len(isDirected) > 0 {
		directed = isDirected[0]
	}
	s := &types.Graph{
		Grabber:    make(map[string]*types.Node),
		IsDirected: directed,
	}

	for _, n := range nodes {
		s.Grabber[n] = &types.Node{
			Id:        n,
			Neighbors: make(map[*types.Node]int),
		}
	}

	for _, e := range edges {
		if len(e) < 2 {
			// bad format continue
			continue
		}

		u, v := e[0], e[1]
		nu, nv := GetNode(s, u), GetNode(s, v)
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

		if s.IsDirected {
			continue // only process explicit edge if directed graph
		}

		nv.Neighbors[nu] = w
		nu.Indegree++
		nv.Outdegree++
	}

	return s
}
