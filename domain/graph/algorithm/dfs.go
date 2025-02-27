package algorithm

import (
	"fmt"

	"github.com/msyamsula/portofolio/domain/graph"
)

func (s *Service) Dfs(g *graph.Service, start *graph.Node) []string {
	s.dfsLog = []string{}
	start.Parent = nil
	for _, n := range g.Grabber {
		n.Visited = false
		n.Parent = nil
	}

	s.dfs(start)
	return s.dfsLog
}

func (s *Service) dfs(u *graph.Node) {
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

func (s *Service) constructPath(end, start *graph.Node) []string {
	ptr := end
	backPath := []string{}
	for ptr != start {
		backPath = append(backPath, ptr.Id)
		ptr = ptr.Parent
	}

	path := []string{}
	if start != nil {
		path = append(path, start.Id)
	}
	for i := len(backPath) - 1; i >= 0; i-- {
		path = append(path, backPath[i])
	}

	return path
}
