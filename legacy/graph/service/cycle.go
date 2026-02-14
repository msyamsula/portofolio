package service

import (
	"fmt"

	"github.com/msyamsula/portofolio/backend-app/graph/types"
)

func (s *service) IsCycle(g *types.Graph) (log []string, cycles [][]string) {
	for _, n := range g.Grabber {
		n.Color = types.White
	}
	s.CycleLog = []string{}
	s.CycleExist = false

	for _, u := range g.Grabber {
		if u.Color == types.Black {
			continue
		}
		if s.isCycle(u, g.IsDirected) {
			break
		}
	}

	cycles = make([][]string, 0)
	for _, p := range s.Cycle {
		cycles = append(cycles, s.constructPath(p.End, p.Start))
	}

	return s.CycleLog, cycles
}

func (s *service) isCycle(u *types.Node, directed bool) bool {
	u.Color = types.Grey
	s.CycleLog = append(s.CycleLog, fmt.Sprintf("grey:%s", u.Id))
	defer func() {
		u.Color = types.Black
		s.CycleLog = append(s.CycleLog, fmt.Sprintf("black:%s", u.Id))
	}()

	for v := range u.Neighbors {
		if u.Parent == v && !directed {
			continue
		}
		s.CycleLog = append(s.CycleLog, fmt.Sprintf("edge:%s:%s", u.Id, v.Id))
		switch v.Color {
		case types.White:
			v.Parent = u
			if s.isCycle(v, directed) {
				return true
			}
		case types.Black:
			// do nothing
		default:
			// cycle detected
			s.CycleLog = append(s.CycleLog, fmt.Sprintf("cycle:%s:%s", u.Id, v.Id))
			s.Cycle = append(s.Cycle, types.CyclePair{
				Start: v,
				End:   u,
			})
			return true
		}
		s.CycleLog = append(s.CycleLog, fmt.Sprintf("deEdge:%s:%s", u.Id, v.Id))
	}

	return false
}

func (s *service) constructPath(end, start *types.Node) []string {
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
