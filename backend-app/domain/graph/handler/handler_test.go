package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/backend-app/domain/graph/dto"
	"github.com/msyamsula/portofolio/backend-app/domain/graph/service"
	"github.com/msyamsula/portofolio/backend-app/mock"
	"github.com/stretchr/testify/suite"
)

// GraphHandlerTestSuite defines the test suite for graph handler
type GraphHandlerTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	mockSvc *mock.MockGraphService
	handler *Handler
	router  *mux.Router
}

func (s *GraphHandlerTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockSvc = mock.NewMockGraphService(s.ctrl)
	s.handler = New(s.mockSvc)
	s.router = mux.NewRouter()
	s.handler.RegisterRoutes(s.router)
}

func (s *GraphHandlerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *GraphHandlerTestSuite) makeRequest(algo string, body dto.SolveRequest, query string) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	path := "/solve/" + algo
	if query != "" {
		path += "?" + query
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GraphHandlerTestSuite) sampleRequest() dto.SolveRequest {
	return dto.SolveRequest{
		Graph: dto.GraphNotation{
			Nodes: []string{"A", "B", "C"},
			Edges: []dto.Edge{
				{From: "A", To: "B", Weight: "1"},
				{From: "B", To: "C", Weight: "1"},
			},
		},
	}
}

// --- DFS test ---

func (s *GraphHandlerTestSuite) TestSolve_DFS() {
	s.mockSvc.EXPECT().DepthFirstSearch(gomock.Any()).Return([]string{"A", "B", "C"})

	rr := s.makeRequest("dfs", s.sampleRequest(), "")
	s.Equal(http.StatusOK, rr.Code)
}

// --- BFS test ---

func (s *GraphHandlerTestSuite) TestSolve_BFS() {
	s.mockSvc.EXPECT().BreadthFirstSearch(gomock.Any()).Return([]string{"A", "B", "C"})

	rr := s.makeRequest("bfs", s.sampleRequest(), "")
	s.Equal(http.StatusOK, rr.Code)
}

// --- Cycle detection test ---

func (s *GraphHandlerTestSuite) TestSolve_Cycle() {
	s.mockSvc.EXPECT().IsCycle(gomock.Any()).Return([]string{"log"}, [][]string{{"A", "B"}})

	rr := s.makeRequest("cycle", s.sampleRequest(), "")
	s.Equal(http.StatusOK, rr.Code)
}

// --- DAG test ---

func (s *GraphHandlerTestSuite) TestSolve_DAG() {
	s.mockSvc.EXPECT().DirectedAcyclicGraph(gomock.Any()).Return([]string{"A", "B", "C"}, true)

	rr := s.makeRequest("dag", s.sampleRequest(), "isDirected=true")
	s.Equal(http.StatusOK, rr.Code)
}

// --- SCC test ---

func (s *GraphHandlerTestSuite) TestSolve_SCC() {
	s.mockSvc.EXPECT().StronglyConnectedComponents(gomock.Any()).
		Return([]string{"log"}, [][]string{{"A", "B"}})

	rr := s.makeRequest("scc", s.sampleRequest(), "isDirected=true")
	s.Equal(http.StatusOK, rr.Code)
}

// --- Articulation Point test ---

func (s *GraphHandlerTestSuite) TestSolve_ArticulationPoint() {
	s.mockSvc.EXPECT().ArticulationPointAndBridge(gomock.Any()).
		Return([]string{"log"}, []string{"B"}, [][]string{{"A", "B"}})

	rr := s.makeRequest("ap", s.sampleRequest(), "")
	s.Equal(http.StatusOK, rr.Code)
}

// --- Eulerian Path test ---

func (s *GraphHandlerTestSuite) TestSolve_Eulerian() {
	s.mockSvc.EXPECT().Eulerian(gomock.Any()).Return([]string{"A", "B", "C"})

	rr := s.makeRequest("ep", s.sampleRequest(), "")
	s.Equal(http.StatusOK, rr.Code)
}

// --- Error cases ---

func (s *GraphHandlerTestSuite) TestSolve_InvalidAlgorithm() {
	rr := s.makeRequest("unknown", s.sampleRequest(), "")
	s.Equal(http.StatusBadRequest, rr.Code)
}

func (s *GraphHandlerTestSuite) TestSolve_InvalidBody() {
	req := httptest.NewRequest(http.MethodPost, "/solve/dfs", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	s.Equal(http.StatusBadRequest, rr.Code)
}

func (s *GraphHandlerTestSuite) TestSolve_IsDirectedQueryParam() {
	// Verify isDirected=true creates a directed graph
	s.mockSvc.EXPECT().DepthFirstSearch(gomock.AssignableToTypeOf(&service.Graph{})).
		DoAndReturn(func(g *service.Graph) []string {
			s.True(g.IsDirected)
			return []string{"A"}
		})

	rr := s.makeRequest("dfs", s.sampleRequest(), "isDirected=true")
	s.Equal(http.StatusOK, rr.Code)
}

func (s *GraphHandlerTestSuite) TestSolve_UndirectedByDefault() {
	s.mockSvc.EXPECT().DepthFirstSearch(gomock.AssignableToTypeOf(&service.Graph{})).
		DoAndReturn(func(g *service.Graph) []string {
			s.False(g.IsDirected)
			return []string{"A"}
		})

	rr := s.makeRequest("dfs", s.sampleRequest(), "")
	s.Equal(http.StatusOK, rr.Code)
}

// --- Constructor & Routes tests ---

func (s *GraphHandlerTestSuite) TestNew_ReturnsHandler() {
	h := New(s.mockSvc)
	s.NotNil(h)
}

func (s *GraphHandlerTestSuite) TestRegisterRoutes() {
	r := mux.NewRouter()
	s.handler.RegisterRoutes(r)

	var match mux.RouteMatch
	req := httptest.NewRequest(http.MethodPost, "/solve/dfs", nil)
	s.True(r.Match(req, &match))
}

// Run the test suite
func TestGraphHandlerSuite(t *testing.T) {
	suite.Run(t, new(GraphHandlerTestSuite))
}
