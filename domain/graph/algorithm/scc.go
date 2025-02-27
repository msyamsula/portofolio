package algorithm

import (
	"sort"

	"github.com/msyamsula/portofolio/domain/graph"
)

func (s *Service) transpose(g *graph.Service) *graph.Service {
	gt := &graph.Service{
		Grabber:    make(map[string]*graph.Node),
		IsDirected: g.IsDirected,
	}

	for id := range g.Grabber {
		gt.Grabber[id] = &graph.Node{
			Id:        id,
			Neighbors: make(map[*graph.Node]int),
		}
	}

	for _, u := range g.Grabber {
		for v, w := range u.Neighbors {
			gt.GetNode(v.Id).Neighbors[gt.GetNode(u.Id)] = w
		}
	}

	return gt
}

func (s *Service) StronglyConnectedComponents(g *graph.Service) (log []string, comp [][]string) {

	nodeList := []*graph.Node{}
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
		if gt.GetNode(n.Id).Visited {
			continue
		}
		s.sccTree = []string{}
		s.getSccTree(gt.GetNode(n.Id))
		comp = append(comp, s.sccTree)
	}

	return []string{}, comp
}

func (s *Service) getSccTree(u *graph.Node) {
	u.Visited = true
	s.sccTree = append(s.sccTree, u.Id)

	for v := range u.Neighbors {
		if v.Visited {
			continue
		}

		s.getSccTree(v)
	}
}

func (s *Service) dfsTimer(u *graph.Node, timer *int) {
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
