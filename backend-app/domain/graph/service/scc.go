package service

import "sort"

// StronglyConnectedComponents finds all SCCs in the graph
func (s *service) StronglyConnectedComponents(g *Graph) (log []string, comp [][]string) {

	nodeList := []*Node{}
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
		s.sccTree = []string{}
		s.getSccTree(GetNode(gt, n.Id))
		comp = append(comp, s.sccTree)
	}

	return []string{}, comp
}

func (s *service) transpose(g *Graph) *Graph {
	gt := &Graph{
		Grabber:    make(map[string]*Node),
		IsDirected: g.IsDirected,
	}

	for id := range g.Grabber {
		gt.Grabber[id] = &Node{
			Id:        id,
			Neighbors: make(map[*Node]int),
		}
	}

	for _, u := range g.Grabber {
		for v, w := range u.Neighbors {
			GetNode(g, v.Id).Neighbors[GetNode(g, u.Id)] = w
		}
	}

	return gt
}

func (s *service) getSccTree(u *Node) {
	u.Visited = true
	s.sccTree = append(s.sccTree, u.Id)

	for v := range u.Neighbors {
		if v.Visited {
			continue
		}

		s.getSccTree(v)
	}
}

func (s *service) dfsTimer(u *Node, timer *int) {
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
