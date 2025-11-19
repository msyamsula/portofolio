package service

import (
	"sort"

	"github.com/msyamsula/portofolio/backend-app/graph/types"
)

func (s *service) transpose(g *types.Graph) *types.Graph {
	gt := &types.Graph{
		Grabber:    make(map[string]*types.Node),
		IsDirected: g.IsDirected,
	}

	for id := range g.Grabber {
		gt.Grabber[id] = &types.Node{
			Id:        id,
			Neighbors: make(map[*types.Node]int),
		}
	}

	for _, u := range g.Grabber {
		for v, w := range u.Neighbors {
			GetNode(g, v.Id).Neighbors[GetNode(g, u.Id)] = w
		}
	}

	return gt
}

func (s *service) StronglyConnectedComponents(g *types.Graph) (log []string, comp [][]string) {

	nodeList := []*types.Node{}
	for _, n := range g.Grabber {
		n.Visited = false
		n.Tout = 0
		n.Tin = 0
		nodeList = append(nodeList, n)
	}

	timer := 0
	for _, n := range g.Grabber {
		if n.Visited {
			continue
		}
		s.dfsTimer(n, &timer)
	}

	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i].Tout > nodeList[j].Tout
	})

	gt := s.transpose(g)

	comp = [][]string{}
	for _, n := range nodeList {
		if GetNode(gt, n.Id).Visited {
			continue
		}
		s.SccTree = []string{}
		s.getSccTree(GetNode(gt, n.Id))
		comp = append(comp, s.SccTree)
	}

	return []string{}, comp
}

func (s *service) getSccTree(u *types.Node) {
	u.Visited = true
	s.SccTree = append(s.SccTree, u.Id)

	for v := range u.Neighbors {
		if v.Visited {
			continue
		}

		s.getSccTree(v)
	}
}

func (s *service) dfsTimer(u *types.Node, timer *int) {
	u.Visited = true
	*timer++
	u.Tin = *timer

	for v := range u.Neighbors {
		if v.Visited {
			continue
		}

		s.dfsTimer(v, timer)
	}

	*timer++
	u.Tout = *timer
}
