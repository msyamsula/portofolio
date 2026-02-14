package handler

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/backend-app/domain/graph/dto"
	"github.com/msyamsula/portofolio/backend-app/domain/graph/service"
	infraHandler "github.com/msyamsula/portofolio/backend-app/infrastructure/http/handler"
	infraLogger "github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
)

// Handler handles HTTP requests for graph algorithms
type Handler struct {
	graphService service.Service
}

// New creates a new graph handler
func New(svc service.Service) *Handler {
	return &Handler{
		graphService: svc,
	}
}

// Solve handles POST /graph/solve/{algo} requests
// @Summary Solve graph algorithm
// @Description Executes the specified graph algorithm on the provided graph
// @Tags graph
// @Accept json
// @Produce json
// @Param algo path string true "Algorithm name (dfs, bfs, cycle, dag, scc, ap, ep)"
// @Param isDirected query string false "Is graph directed"
// @Param body body handler.SolveRequest true "Graph notation"
// @Success 200 {object} handler.SolveResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /graph/solve/{algo} [post]
func (h *Handler) Solve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	infraLogger.Info("graph solve request started", map[string]any{
		"method": r.Method,
		"path":   r.URL.Path,
		"query":  r.URL.RawQuery,
	})

	// Create child span for handler logic
	tracer := otel.Tracer("graph")
	ctx, span := tracer.Start(ctx, "handler.solve")
	defer span.End()

	// Get algorithm from path variable
	algo := infraHandler.PathVar(r, "algo")
	if algo == "" {
		infraLogger.Warn("graph solve request missing algorithm", map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.SetStatus(codes.Error, "algorithm is required")
		_ = infraHandler.BadRequest(w, "algorithm is required")
		return
	}

	// Parse request body
	var req dto.SolveRequest
	if err := infraHandler.BindJSON(r, &req); err != nil {
		infraLogger.WarnError("graph solve request invalid body", err, map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request body")
		_ = infraHandler.BadRequest(w, "invalid request body")
		return
	}

	// Get isDirected query parameter
	isDirected := infraHandler.QueryParam(r, "isDirected") == "true"

	// Add attributes to span
	span.SetAttributes(
		attribute.String("graph.algorithm", algo),
		attribute.Bool("graph.is_directed", isDirected),
		attribute.Int("graph.node_count", len(req.Graph.Nodes)),
		attribute.Int("graph.edge_count", len(req.Graph.Edges)),
	)

	// Construct graph from notation
	edges := [][]string{}
	for _, e := range req.Graph.Edges {
		edges = append(edges, []string{e.From, e.To, e.Weight})
	}
	graph := service.NewGraph(req.Graph.Nodes, edges, isDirected)

	// Execute algorithm based on path variable
	var result dto.AlgorithmResult
	var err error
	switch algo {
	case "dfs":
		result.Log = h.graphService.DepthFirstSearch(graph)
	case "bfs":
		result.Log = h.graphService.BreadthFirstSearch(graph)
	case "cycle":
		result.Log, result.Cycles = h.graphService.IsCycle(graph)
	case "dag":
		result.Path, result.Acyclic = h.graphService.DirectedAcyclicGraph(graph)
	case "scc":
		result.Log, result.Scc = h.graphService.StronglyConnectedComponents(graph)
	case "ap":
		result.Log, result.Ap, result.Bridge = h.graphService.ArticulationPointAndBridge(graph)
	case "ep":
		result.Path = h.graphService.Eulerian(graph)
	default:
		err = http.ErrNotSupported
	}

	if err != nil {
		infraLogger.WarnError("graph solve request invalid algorithm", err, map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"algorithm":   algo,
			"duration_ms": time.Since(start).Milliseconds(),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid algorithm")
		_ = infraHandler.BadRequest(w, "invalid algorithm")
		return
	}

	// Return response
	resp := dto.SolveResponse(result)
	_ = infraHandler.OK(w, resp)

	infraLogger.Info("graph solve request completed", map[string]any{
		"method":        r.Method,
		"path":          r.URL.Path,
		"algorithm":     algo,
		"is_directed":   isDirected,
		"node_count":    len(req.Graph.Nodes),
		"edge_count":    len(req.Graph.Edges),
		"duration_ms":   time.Since(start).Milliseconds(),
	})
}

// RegisterRoutes registers all graph handler routes
func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/graph/solve/{algo}", h.Solve).Methods("POST")
}
