package service

import (
	"fmt"
)

func (s *Algorithm) DepthFirstSearch(g *Graph) []string {
	for _, n := range g.Grabber {
		n.Visited = false
		n.Parent = nil
	}
	s.dfsLog = []string{}

	for _, n := range g.Grabber {
		if n.Visited {
			continue
		}
		s.dfs(n)

	}
	return s.dfsLog
}

func (s *Algorithm) dfs(u *Node) {
	u.Visited = true
	s.dfsLog = append(s.dfsLog, fmt.Sprintf("node:%s", u.Id))

	for v := range u.Neighbors {
		if v.Visited {
			continue
		} else {
			v.Parent = u
			s.dfsLog = append(s.dfsLog, fmt.Sprintf("edge:%s:%s", u.Id, v.Id))
			s.dfs(v)
			s.dfsLog = append(s.dfsLog, fmt.Sprintf("deEdge:%s:%s", u.Id, v.Id))
		}
	}

	s.dfsLog = append(s.dfsLog, fmt.Sprintf("deNode:%s", u.Id))
}
