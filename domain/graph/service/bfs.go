package service

import (
	"fmt"
)

func (s *Algorithm) BreadthFirstSearch(g *Graph) []string {
	// clean up previous work first
	for _, n := range g.Grabber {
		n.Visited = false
		n.Parent = nil
	}
	s.bfsLog = []string{}

	for _, n := range g.Grabber {
		if n.Visited {
			continue
		}
		s.bfs(n)
	}
	return s.bfsLog
}

func (s *Algorithm) bfs(start *Node) {
	queue := make([]*Node, 0)
	queue = append(queue, start)
	start.Parent = nil
	start.Visited = true
	s.bfsLog = append(s.bfsLog, fmt.Sprintf("node:%s", start.Id))

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		s.bfsLog = append(s.bfsLog, fmt.Sprintf("bold:%s", u.Id))

		for v := range u.Neighbors {
			if v.Visited {
				continue
			} else {
				s.bfsLog = append(s.bfsLog, fmt.Sprintf("edge:%s:%s", u.Id, v.Id))
				s.bfsLog = append(s.bfsLog, fmt.Sprintf("node:%s", v.Id))
			}

		}
		s.bfsLog = append(s.bfsLog, fmt.Sprintf("deBold:%s", u.Id))
		for v := range u.Neighbors {
			if v.Visited {
				continue
			} else {
				v.Visited = true
				v.Parent = u
				queue = append(queue, v)
				s.bfsLog = append(s.bfsLog, fmt.Sprintf("deEdge:%s:%s", u.Id, v.Id))
			}

		}
		s.bfsLog = append(s.bfsLog, fmt.Sprintf("deNode:%s", u.Id))
	}
}
