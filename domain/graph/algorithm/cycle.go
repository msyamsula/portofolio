package algorithm

import (
	"fmt"

	"github.com/msyamsula/portofolio/domain/graph"
)

func (s *Service) IsCycle(g *graph.Service) (log []string, cycles [][]string) {
	for _, n := range g.Grabber {
		n.Color = graph.White
	}
	s.cycleLog = []string{}
	s.cycleExist = false

	for _, u := range g.Grabber {
		if u.Color == graph.Black {
			continue
		}
		s.isCycle(u)
	}

	cycles = make([][]string, 0)
	for _, p := range s.cycle {
		cycles = append(cycles, s.constructPath(p.end, p.start))
	}

	return s.cycleLog, cycles
}

func (s *Service) isCycle(u *graph.Node) {
	u.Color = graph.Grey
	s.cycleLog = append(s.cycleLog, fmt.Sprintf("grey:%s", u.Id))

	for v := range u.Neighbors {
		if u.Parent == v {
			continue
		}
		s.cycleLog = append(s.cycleLog, fmt.Sprintf("edge:%s:%s", u.Id, v.Id))
		if v.Color == graph.White {
			v.Parent = u
			s.isCycle(v)
		} else if v.Color == graph.Black {
			// do nothing
		} else {
			// cycle detected
			s.cycleLog = append(s.cycleLog, fmt.Sprintf("cycle:%s:%s", u.Id, v.Id))
			s.cycle = append(s.cycle, CyclePair{
				start: v,
				end:   u,
			})
			s.cycleExist = true
		}
		s.cycleLog = append(s.cycleLog, fmt.Sprintf("deEdge:%s:%s", u.Id, v.Id))
	}

	u.Color = graph.Black
	s.cycleLog = append(s.cycleLog, fmt.Sprintf("black:%s", u.Id))
}
