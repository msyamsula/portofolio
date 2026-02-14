package handler

import (
	"net/http"

	"github.com/msyamsula/portofolio/backend-app/graph/service"
	"github.com/msyamsula/portofolio/backend-app/graph/types"
)

type Handler interface {
	Solve(w http.ResponseWriter, r *http.Request)
}

type Config struct {
	Service service.Service
}

func NewHandler(cfg Config) Handler {
	return &handler{
		graph: new(types.Graph), // initialize during runtime
		svc:   cfg.Service,
	}
}
