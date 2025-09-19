package service

import (
	"fmt"
)

func (s *Algorithm) ArticulationPointAndBridge(g *Graph) (log []string, id []string, bridge [][]string) {
	for _, n := range g.Grabber {
		n.Parent = nil
		n.Color = White
		n.Tin = 0
		n.Low = 0
	}
	s.apLog = []string{}
	s.apId = []string{}
	s.bridge = [][]string{}

	timer := 0
	for _, n := range g.Grabber {
		if n.Color == Black {
			continue
		}
		rootChild := 0
		s.ap(n, &timer, n, &rootChild)
	}

	return s.apLog, s.apId, s.bridge
}

func (s *Algorithm) ap(u *Node, timer *int, root *Node, child *int) {
	*timer++
	u.Color = Grey
	u.Tin = *timer
	s.apLog = append(s.apLog, fmt.Sprintf("grey:%s", u.Id))

	u.Low = u.Tin
	s.apLog = append(s.apLog, fmt.Sprintf("label:%s:%d:%d", u.Id, u.Tin, u.Low))
	for v := range u.Neighbors {
		if v == u.Parent {
			continue
		}
		s.apLog = append(s.apLog, fmt.Sprintf("edge:%s:%s", u.Id, v.Id))
		switch v.Color {
		case Grey:
			u.Low = min(u.Low, v.Tin)
			s.apLog = append(s.apLog, fmt.Sprintf("label:%s:%d:%d", u.Id, u.Tin, u.Low))
		case Black:
			u.Low = min(u.Low, v.Low)
			s.apLog = append(s.apLog, fmt.Sprintf("label:%s:%d:%d", u.Id, u.Tin, u.Low))
		default:
			if u == root {
				*child++
			}
			v.Parent = u
			s.ap(v, timer, root, child)
			u.Low = min(u.Low, v.Low)
			s.apLog = append(s.apLog, fmt.Sprintf("label:%s:%d:%d", u.Id, u.Tin, u.Low))
			if v.Low > u.Tin {
				s.apLog = append(s.apLog, fmt.Sprintf("bridge:%s:%s", u.Id, v.Id))
				s.bridge = append(s.bridge, []string{u.Id, v.Id})
			}
			if v.Low >= u.Tin && u != root {
				s.apLog = append(s.apLog, fmt.Sprintf("ap:%s", u.Id))
				u.IsArticulationPoint = true
				s.apId = append(s.apId, u.Id)
			}
		}

		s.apLog = append(s.apLog, fmt.Sprintf("deEdge:%s:%s", u.Id, v.Id))
	}

	if *child > 1 && u == root {
		s.apLog = append(s.apLog, fmt.Sprintf("ap:%s", u.Id))
		u.IsArticulationPoint = true
		s.apId = append(s.apId, u.Id)
	}

	u.Color = Black
	s.apLog = append(s.apLog, fmt.Sprintf("white:%s", u.Id))
}
