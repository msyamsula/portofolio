package graph

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/msyamsula/portofolio/binary/telemetry"
	graphhttp "github.com/msyamsula/portofolio/domain/graph/http"
)

func Run(r *mux.Router) {
	appName := "graph"

	// load env
	godotenv.Load(".env")

	// initialize instrumentation
	telemetry.InitializeTelemetryTracing(appName, os.Getenv("JAEGER_HOST"))

	// graph handler
	graphHandler := graphhttp.NewHandler()

	// graph
	r.HandleFunc("/graph/{algo}", http.HandlerFunc(graphHandler.InitGraph(
		http.HandlerFunc(graphHandler.Algorithm)),
	)).Methods(http.MethodGet)
}
