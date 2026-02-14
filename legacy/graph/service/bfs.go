package service

import (
	"fmt"

	"github.com/msyamsula/portofolio/backend-app/graph/types"
)

func (s *service) BreadthFirstSearch(g *types.Graph) []string {
	// clean up previous work first
	for _, n := range g.Grabber {
		n.Visited = false
		n.Parent = nil
	}
	s.BfsLog = []string{}

	for _, n := range g.Grabber {
		if n.Visited {
			continue
		}
		s.bfs(n)
	}
	return s.BfsLog
}

func (s *service) bfs(start *types.Node) {
	queue := make([]*types.Node, 0)
	queue = append(queue, start)
	start.Parent = nil
	start.Visited = true
	s.BfsLog = append(s.BfsLog, fmt.Sprintf("node:%s", start.Id))

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		s.BfsLog = append(s.BfsLog, fmt.Sprintf("bold:%s", u.Id))

		for v := range u.Neighbors {
			if v.Visited {
				continue
			} else {
				s.BfsLog = append(s.BfsLog, fmt.Sprintf("edge:%s:%s", u.Id, v.Id))
				s.BfsLog = append(s.BfsLog, fmt.Sprintf("node:%s", v.Id))
			}

		}
		s.BfsLog = append(s.BfsLog, fmt.Sprintf("deBold:%s", u.Id))
		for v := range u.Neighbors {
			if v.Visited {
				continue
			} else {
				v.Visited = true
				v.Parent = u
				queue = append(queue, v)
				s.BfsLog = append(s.BfsLog, fmt.Sprintf("deEdge:%s:%s", u.Id, v.Id))
			}

		}
		s.BfsLog = append(s.BfsLog, fmt.Sprintf("deNode:%s", u.Id))
	}
}
