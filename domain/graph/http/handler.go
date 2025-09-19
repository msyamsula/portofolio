package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/domain/graph/service"
)

type Handler struct {
	graph *service.Graph
}

type graphNotation struct {
	Nodes []string `json:"nodes,omitempty"`
	Edges []Edge   `json:"edges,omitempty"`
}

type Edge struct {
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	Weight string `json:"weight,omitempty"`
}

// InitGraph is a middleware, it create graph before executing an algorithm
func (s *Handler) InitGraph(next http.Handler) http.HandlerFunc {
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
		body := graphNotation{}
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

		s.graph = service.NewGraph(body.Nodes, edges, directed)
		next.ServeHTTP(w, r)
	}

}

type AlgoResult struct {
	Log     []string   `json:"log"`
	Path    []string   `json:"path"`
	Cycles  [][]string `json:"cycles"`
	Acyclic bool       `json:"acyclic"`
	Scc     [][]string `json:"scc"`
	Ap      []string   `json:"ap"`
	Bridge  [][]string `json:"bridge"`
}

func (s *Handler) Algorithm(w http.ResponseWriter, r *http.Request) {
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

	algorithm := service.New()
	result := AlgoResult{}
	switch algo {
	case "dfs":
		log := algorithm.DepthFirstSearch(s.graph)
		result = AlgoResult{
			Log: log,
		}
	case "bfs":
		log := algorithm.BreadthFirstSearch(s.graph)
		result = AlgoResult{
			Log: log,
		}
	case "cycle":
		log, cycles := algorithm.IsCycle(s.graph)
		result = AlgoResult{
			Log:    log,
			Cycles: cycles,
		}
	case "dag":
		path, acyclic := algorithm.DirectedAcyclicGraph(s.graph)
		result = AlgoResult{
			Path:    path,
			Acyclic: acyclic,
		}
	case "scc":
		log, scc := algorithm.StronglyConnectedComponents(s.graph)
		result = AlgoResult{
			Log: log,
			Scc: scc,
		}
	case "ap":
		log, apId, bridge := algorithm.ArticulationPointAndBridge(s.graph)
		result = AlgoResult{
			Log:    log,
			Ap:     apId,
			Bridge: bridge,
		}
	case "ep":
		eulerPath := algorithm.Eulerian(s.graph)
		result = AlgoResult{
			Path: eulerPath,
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("algo is not defined"))
		return
	}

	json.NewEncoder(w).Encode(result)
}

func NewHandler() *Handler {
	return &Handler{}
}
