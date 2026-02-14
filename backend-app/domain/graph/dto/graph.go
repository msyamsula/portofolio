package dto

// Edge represents a graph edge with direction and weight
type Edge struct {
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	Weight string `json:"weight,omitempty"`
}

// GraphNotation represents the graph notation from HTTP request
type GraphNotation struct {
	Nodes []string `json:"nodes,omitempty"`
	Edges []Edge   `json:"edges,omitempty"`
}

// AlgorithmResult represents the result of graph algorithm execution
type AlgorithmResult struct {
	Log     []string   `json:"log"`
	Path    []string   `json:"path"`
	Cycles  [][]string `json:"cycles"`
	Acyclic bool       `json:"acyclic"`
	Scc     [][]string `json:"scc"`
	Ap      []string   `json:"ap"`
	Bridge  [][]string `json:"bridge"`
}

// SolveRequest represents the request to solve a graph algorithm
type SolveRequest struct {
	Graph GraphNotation `json:"graph"`
}

// SolveResponse represents the response from solving a graph algorithm
type SolveResponse AlgorithmResult
