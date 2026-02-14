package service

import "fmt"

// IsCycle checks if the graph contains cycles
func (s *service) IsCycle(g *Graph) (log []string, cycles [][]string) {
	for _, n := range g.Grabber {
		n.Color = White
	}
	s.cycleLog = []string{}

	for _, u := range g.Grabber {
		if u.Color == Black {
			continue
		}
		if s.isCycle(u, g.IsDirected) {
			break
		}
	}

	cycles = make([][]string, 0)
	for _, p := range s.cycle {
		cycles = append(cycles, s.constructPath(p.End, p.Start))
	}

	return s.cycleLog, cycles
}

func (s *service) isCycle(u *Node, directed bool) bool {
	u.Color = Grey
	s.cycleLog = append(s.cycleLog, fmt.Sprintf("grey:%s", u.Id))
	defer func() {
		u.Color = Black
		s.cycleLog = append(s.cycleLog, fmt.Sprintf("black:%s", u.Id))
	}()

	for v := range u.Neighbors {
		if u.Parent == v && !directed {
			continue
		}
		s.cycleLog = append(s.cycleLog, fmt.Sprintf("edge:%s:%s", u.Id, v.Id))
		switch v.Color {
		case White:
			v.Parent = u
			if s.isCycle(v, directed) {
				return true
			}
		case Black:
			// do nothing
		default:
			// cycle detected
			s.cycleLog = append(s.cycleLog, fmt.Sprintf("cycle:%s:%s", u.Id, v.Id))
			s.cycle = append(s.cycle, CyclePair{
				Start: v,
				End:   u,
			})
			return true
		}
		s.cycleLog = append(s.cycleLog, fmt.Sprintf("deEdge:%s:%s", u.Id, v.Id))
	}

	return false
}

func (s *service) constructPath(end, start *Node) []string {
	ptr := end
	backPath := []string{}
	for ptr != start {
		backPath = append(backPath, ptr.Id)
		ptr = ptr.Parent
	}

	backPath = append(backPath, start.Id)

	path := []string{}
	for i := len(backPath) - 1; i >= 0; i-- {
		path = append(path, backPath[i])
	}

	return path
}
