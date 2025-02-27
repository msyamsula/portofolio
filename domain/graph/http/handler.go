package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/domain/graph"
	"github.com/msyamsula/portofolio/domain/graph/algorithm"
)

type Service struct {
	graph *graph.Service
}

type dfsBody struct {
	Nodes []string `json:"nodes,omitempty"`
	Edges []Edge   `json:"edges,omitempty"`
}

type Edge struct {
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	Weight string `json:"weight,omitempty"`
}

func (s *Service) InitGraph(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		q := r.URL.Query()
		directed := q.Get("isDirected") == "true"
		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad body"))
			return
		}
		body := dfsBody{}
		err = json.Unmarshal(b, &body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad body"))
			return
		}

		edges := [][]string{}
		for _, e := range body.Edges {
			edges = append(edges, []string{e.From, e.To, e.Weight})
		}

		s.graph = graph.New(body.Nodes, edges, directed)
		next.ServeHTTP(w, r)
	}

}

type AlgoResult struct {
	Log     []string   `json:"log"`
	Path    []string   `json:"path"`
	Cycles  [][]string `json:"cycles"`
	Acyclic bool       `json:"acyclic"`
	Scc     [][]string `json:"scc"`
}

func (s *Service) Algorithm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.graph == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("graph is not initiated"))
		return
	}
	pathVariable := mux.Vars(r)
	algo := ""
	if pathVariable != nil && pathVariable["algo"] != "" {
		algo = pathVariable["algo"]
	}

	machine := algorithm.New()
	result := AlgoResult{}
	switch algo {
	case "dfs", "bfs":
		var log []string
		if algo == "dfs" {
			log = machine.DepthFirstSearch(s.graph)
		} else {
			log = machine.BreadthFirstSearch(s.graph)
		}
		result = AlgoResult{
			Log: log,
		}
	case "cycle":
		log, cycles := machine.IsCycle(s.graph)
		result = AlgoResult{
			Log:    log,
			Cycles: cycles,
		}
	case "dag":
		path, acyclic := machine.DirectedAcyclicGraph(s.graph)
		result = AlgoResult{
			Path:    path,
			Acyclic: acyclic,
		}
	case "scc":
		log, scc := machine.StronglyConnectedComponents(s.graph)
		result = AlgoResult{
			Log: log,
			Scc: scc,
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("algo is not defined"))
		return
	}

	resp, err := json.Marshal(result)
	fmt.Println(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(resp)
}
