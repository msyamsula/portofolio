package service

// Eulerian finds an Eulerian path or cycle in the graph
func (s *service) Eulerian(g *Graph) (path []string) {
	for _, n := range g.Grabber {
		n.Visited = false
	}
	s.eulerPath = []string{}

	var start *Node
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

	s.dfsTree = []string{}
	s.ep(start, g.IsDirected)

	// reverse path
	n := len(s.eulerPath)
	for i := 0; i <= (n-1)/2; i++ {
		s.eulerPath[i], s.eulerPath[n-1-i] = s.eulerPath[n-1-i], s.eulerPath[i]
	}

	return s.eulerPath
}

func (s *service) ep(u *Node, directed bool) {
	s.dfsTree = append(s.dfsTree, u.Id)

	for v := range u.Neighbors {
		delete(u.Neighbors, v)
		if !directed {
			delete(v.Neighbors, u)
		}
		s.ep(v, directed)
	}

	back := s.dfsTree[len(s.dfsTree)-1]
	s.dfsTree = s.dfsTree[:len(s.dfsTree)-1]
	s.eulerPath = append(s.eulerPath, back)
}
