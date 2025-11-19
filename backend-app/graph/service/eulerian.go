package service

import "github.com/msyamsula/portofolio/backend-app/graph/types"

func (s *service) Eulerian(g *types.Graph) (path []string) {
	for _, n := range g.Grabber {
		n.Visited = false
	}
	s.EulerPath = []string{}

	var start *types.Node
	if !g.IsDirected {
		var odd, even int
		for _, n := range g.Grabber {
			if n.Indegree%2 == 0 {
				even++
			} else {
				odd++
				start = n
			}
		}

		if odd == 0 {
			// cycle
			for _, n := range g.Grabber {
				// choose random node as a start
				start = n
				break
			}
		} else if odd == 2 {
			// path
		} else {
			// eulerian doesn't exist
			return []string{}
		}
	} else {
		var pOne, nOne, zero int
		for _, n := range g.Grabber {
			degree := n.Outdegree - n.Indegree
			switch degree {
			case -1:
				nOne++
			case 1:
				pOne++
				start = n
			case 0:
				zero++
			default:
				// if exist then no path or cycle
				return []string{}
			}
		}

		if pOne == 1 && nOne == 1 {
			// path, start node already assign
		} else if pOne == 0 && nOne == 0 {
			// cycle,  grab random node for start
			for _, n := range g.Grabber {
				start = n
				break
			}
		} else {
			// path/cycle does not exist
			return []string{}
		}
	}

	s.DfsTree = []string{}
	s.ep(start, g.IsDirected)

	// reverse the path
	n := len(s.EulerPath)
	for i := 0; i <= (n-1)/2; i++ {
		s.EulerPath[i], s.EulerPath[n-1-i] = s.EulerPath[n-1-i], s.EulerPath[i]
	}

	return s.EulerPath
}

func (s *service) ep(u *types.Node, directed bool) {
	s.DfsTree = append(s.DfsTree, u.Id)

	for v := range u.Neighbors {
		delete(u.Neighbors, v)
		if !directed {
			delete(v.Neighbors, u)
		}
		s.ep(v, directed)
	}

	back := s.DfsTree[len(s.DfsTree)-1]
	s.DfsTree = s.DfsTree[:len(s.DfsTree)-1]
	s.EulerPath = append(s.EulerPath, back)
}
