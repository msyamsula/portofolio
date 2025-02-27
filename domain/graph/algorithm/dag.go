package algorithm

import (
	"github.com/msyamsula/portofolio/domain/graph"
)

func (s *Service) Dag(g *graph.Service) (path []string, acyclic bool) {
	for _, n := range g.Grabber {
		n.Visited = false
	}
	s.dagPath = []string{}
	if g.IsDirected {
		// use kahn
		startNodes := []*graph.Node{}
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

		return s.dagPath, len(s.dagPath) == len(g.Grabber)
	}

	// use cycle check and dfs tree for undirected graph
	_, c := s.IsCycle(g)
	if len(c) > 0 {
		return []string{}, false
	}

	s.dfsLog = []string{}
	for _, n := range g.Grabber {
		if n.Visited {
			continue
		}

		s.dag(g, n)
	}

	// reverse s.dagPath
	for i := 0; i <= (len(s.dagPath)-1)/2; i++ {
		s.dagPath[i], s.dagPath[len(s.dagPath)-1-i] = s.dagPath[len(s.dagPath)-1], s.dagPath[i]
	}

	return s.dagPath, true
}

func (s *Service) dag(g *graph.Service, u *graph.Node) {
	u.Visited = true
	s.dfsLog = append(s.dfsLog, u.Id)

	for v := range u.Neighbors {
		if v.Visited {
			continue
		}
		s.dag(g, v)
	}

	s.dfsLog = s.dfsLog[:len(s.dfsLog)-1]
	s.dagPath = append(s.dagPath, u.Id)
}

func (s *Service) kahn(nodes []*graph.Node) {
	queue := nodes

	s.dagPath = []string{}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		s.dagPath = append(s.dagPath, u.Id)

		for v := range u.Neighbors {
			v.In--
			if v.In == 0 {
				queue = append(queue, v)
			}
		}
	}

}
