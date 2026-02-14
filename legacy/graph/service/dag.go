package service

import "github.com/msyamsula/portofolio/backend-app/graph/types"

func (s *service) DirectedAcyclicGraph(g *types.Graph) (path []string, acyclic bool) {
	for _, n := range g.Grabber {
		n.Visited = false
	}
	s.DagPath = []string{}
	if g.IsDirected {
		// use kahn
		startNodes := []*types.Node{}
		for _, n := range g.Grabber {
			if n.Indegree == 0 {
				startNodes = append(startNodes, n)
			}

			n.In = n.Indegree
			n.Out = n.Outdegree
		}

		if len(startNodes) == 0 {
			return []string{}, false
		}

		s.kahn(startNodes)

		return s.DagPath, len(s.DagPath) == len(g.Grabber)
	}

	// use cycle check and dfs tree for undirected graph
	_, c := s.IsCycle(g)
	if len(c) > 0 {
		return []string{}, false
	}

	s.DfsLog = []string{}
	s.DagPath = []string{}
	for _, n := range g.Grabber {
		if n.Visited {
			continue
		}

		s.dag(n)
		// s.dagPath = append(s.dagPath, n.Id)
	}

	// reverse s.dagPath
	for i := 0; i <= (len(s.DagPath)-1)/2; i++ {
		s.DagPath[i], s.DagPath[len(s.DagPath)-1-i] = s.DagPath[len(s.DagPath)-1-i], s.DagPath[i]
	}

	return s.DagPath, true
}

func (s *service) dag(u *types.Node) {
	u.Visited = true
	s.DfsLog = append(s.DfsLog, u.Id)

	for v := range u.Neighbors {
		if v.Visited {
			continue
		}
		s.dag(v)
	}

	s.DfsLog = s.DfsLog[:len(s.DfsLog)-1]
	s.DagPath = append(s.DagPath, u.Id)
}

func (s *service) kahn(nodes []*types.Node) {
	queue := nodes

	s.DagPath = []string{}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		s.DagPath = append(s.DagPath, u.Id)

		for v := range u.Neighbors {
			v.In--
			if v.In == 0 {
				queue = append(queue, v)
			}
		}
	}

}
