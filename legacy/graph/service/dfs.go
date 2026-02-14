package service

import (
	"fmt"

	"github.com/msyamsula/portofolio/backend-app/graph/types"
)

func (s *service) DepthFirstSearch(g *types.Graph) []string {
	for _, n := range g.Grabber {
		n.Visited = false
		n.Parent = nil
	}
	s.DfsLog = []string{}

	for _, n := range g.Grabber {
		if n.Visited {
			continue
		}
		s.dfs(n)

	}
	return s.DfsLog
}

func (s *service) dfs(u *types.Node) {
	u.Visited = true
	s.DfsLog = append(s.DfsLog, fmt.Sprintf("node:%s", u.Id))

	for v := range u.Neighbors {
		if v.Visited {
			continue
		} else {
			v.Parent = u
			s.DfsLog = append(s.DfsLog, fmt.Sprintf("edge:%s:%s", u.Id, v.Id))
			s.dfs(v)
			s.DfsLog = append(s.DfsLog, fmt.Sprintf("deEdge:%s:%s", u.Id, v.Id))
		}
	}

	s.DfsLog = append(s.DfsLog, fmt.Sprintf("deNode:%s", u.Id))
}
