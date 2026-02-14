package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/backend-app/graph/service"
	"github.com/msyamsula/portofolio/backend-app/graph/types"
)

type handler struct {
	graph *types.Graph
	svc   service.Service
}

func (s *handler) Solve(w http.ResponseWriter, r *http.Request) {
	var notation types.GraphNotation
	var err error

	// get graph notation
	if notation, err = s.getGraphNotation(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// construct graph
	q := r.URL.Query()
	isDirected := q.Get("isDirected") == "true"
	s.graph = s.constructGraph(notation, isDirected)

	pathVariable := mux.Vars(r)
	algo := ""
	if pathVariable != nil && pathVariable["algo"] != "" {
		algo = pathVariable["algo"]
	}
	// solve based on algo
	var result types.AlgorithmResult
	if result, err = s.solve(algo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(&result)
}

func (s *handler) getGraphNotation(w http.ResponseWriter, r *http.Request) (types.GraphNotation, error) {
	w.Header().Set("Content-Type", "application/json")
	body := types.GraphNotation{}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad body"))
		return body, err
	}

	return body, nil
}

func (s *handler) constructGraph(notation types.GraphNotation, isDirected bool) *types.Graph {

	edges := [][]string{}
	for _, e := range notation.Edges {
		edges = append(edges, []string{e.From, e.To, e.Weight})
	}

	return service.NewGraph(notation.Nodes, edges, isDirected)
}

func (s *handler) solve(algo string) (types.AlgorithmResult, error) {
	if s.graph == nil {
		return types.AlgorithmResult{}, errors.New("graph is not initiated")
	}

	var result types.AlgorithmResult
	var err error
	switch algo {
	case "dfs":
		result.Log = s.svc.DepthFirstSearch(s.graph)
	case "bfs":
		result.Log = s.svc.BreadthFirstSearch(s.graph)
	case "cycle":
		result.Log, result.Cycles = s.svc.IsCycle(s.graph)
	case "dag":
		result.Path, result.Acyclic = s.svc.DirectedAcyclicGraph(s.graph)
	case "scc":
		result.Log, result.Scc = s.svc.StronglyConnectedComponents(s.graph)
	case "ap":
		result.Log, result.Ap, result.Bridge = s.svc.ArticulationPointAndBridge(s.graph)
	case "ep":
		result.Path = s.svc.Eulerian(s.graph)
	default:
		err = errors.New("algo is not defined")
	}

	return result, err

}
