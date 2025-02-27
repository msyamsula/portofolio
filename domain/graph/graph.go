package graph

import "strconv"

type Color int

const (
	White Color = iota
	Grey
	Black
)

type Node struct {
	Id        string
	Neighbors map[*Node]int
	Visited   bool
	Color     Color
	Parent    *Node
}

type Service struct {
	Grabber    map[string]*Node
	isDirected bool
}

func (s *Service) GetNode(id string) *Node {
	return s.Grabber[id]
}

func New(nodes []string, edges [][]string, isDirected ...bool) *Service {
	s := &Service{
		Grabber: make(map[string]*Node),
	}

	for _, n := range nodes {
		s.Grabber[n] = &Node{
			Id:        n,
			Neighbors: make(map[*Node]int),
		}
	}

	for _, e := range edges {
		if len(e) < 2 {
			// bad format continue
			continue
		}

		u, v := e[0], e[1]
		nu, nv := s.GetNode(u), s.GetNode(v)
		if nu == nil || nv == nil {
			continue // bad format
		}

		w := 1
		if len(e) >= 3 {
			var err error
			w, err = strconv.Atoi(e[2])
			if err != nil {
				// bad format continue
				w = 1
			}
		}

		if nu.Neighbors == nil {
			nu.Neighbors = make(map[*Node]int)
		}

		nu.Neighbors[nv] = w

		directed := false
		if len(isDirected) > 0 {
			directed = isDirected[0]
		}

		if directed {
			continue // only process explicit edge if directed graph
		}

		if nv.Neighbors == nil {
			nv.Neighbors = make(map[*Node]int)
		}
		nv.Neighbors[nu] = w
	}

	return s
}
