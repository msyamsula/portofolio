package service

import (
	"fmt"

	"github.com/msyamsula/portofolio/backend-app/graph/types"
)

type service struct {
	types.Algorithm
}

func (s *service) ArticulationPointAndBridge(g *types.Graph) (log []string, id []string, bridge [][]string) {
	for _, n := range g.Grabber {
		n.Parent = nil
		n.Color = types.White
		n.Tin = 0
		n.Low = 0
	}
	s.ApLog = []string{}
	s.ApId = []string{}
	s.Bridge = [][]string{}

	timer := 0
	for _, n := range g.Grabber {
		if n.Color == types.Black {
			continue
		}
		rootChild := 0
		s.ap(n, &timer, n, &rootChild)
	}

	return s.ApLog, s.ApId, s.Bridge
}

func (s *service) ap(u *types.Node, timer *int, root *types.Node, child *int) {
	*timer++
	u.Color = types.Grey
	u.Tin = *timer
	s.ApLog = append(s.ApLog, fmt.Sprintf("grey:%s", u.Id))

	u.Low = u.Tin
	s.ApLog = append(s.ApLog, fmt.Sprintf("label:%s:%d:%d", u.Id, u.Tin, u.Low))
	for v := range u.Neighbors {
		if v == u.Parent {
			continue
		}
		s.ApLog = append(s.ApLog, fmt.Sprintf("edge:%s:%s", u.Id, v.Id))
		switch v.Color {
		case types.Grey:
			u.Low = min(u.Low, v.Tin)
			s.ApLog = append(s.ApLog, fmt.Sprintf("label:%s:%d:%d", u.Id, u.Tin, u.Low))
		case types.Black:
			u.Low = min(u.Low, v.Low)
			s.ApLog = append(s.ApLog, fmt.Sprintf("label:%s:%d:%d", u.Id, u.Tin, u.Low))
		default:
			if u == root {
				*child++
			}
			v.Parent = u
			s.ap(v, timer, root, child)
			u.Low = min(u.Low, v.Low)
			s.ApLog = append(s.ApLog, fmt.Sprintf("label:%s:%d:%d", u.Id, u.Tin, u.Low))
			if v.Low > u.Tin {
				s.ApLog = append(s.ApLog, fmt.Sprintf("bridge:%s:%s", u.Id, v.Id))
				s.Bridge = append(s.Bridge, []string{u.Id, v.Id})
			}
			if v.Low >= u.Tin && u != root {
				s.ApLog = append(s.ApLog, fmt.Sprintf("ap:%s", u.Id))
				u.IsArticulationPoint = true
				s.ApId = append(s.ApId, u.Id)
			}
		}

		s.ApLog = append(s.ApLog, fmt.Sprintf("deEdge:%s:%s", u.Id, v.Id))
	}

	if *child > 1 && u == root {
		s.ApLog = append(s.ApLog, fmt.Sprintf("ap:%s", u.Id))
		u.IsArticulationPoint = true
		s.ApId = append(s.ApId, u.Id)
	}

	u.Color = types.Black
	s.ApLog = append(s.ApLog, fmt.Sprintf("white:%s", u.Id))
}
