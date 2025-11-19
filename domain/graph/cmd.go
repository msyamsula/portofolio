package graph

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	graphhttp "github.com/msyamsula/portofolio/domain/graph/http"
)

func Run(r *mux.Router) {

	// load env
	godotenv.Load(".env")

	// graph handler
	graphHandler := graphhttp.NewHandler()

	// graph
	r.HandleFunc("/graph/{algo}", http.HandlerFunc(graphHandler.InitGraph(
		http.HandlerFunc(graphHandler.Algorithm)),
	)).Methods(http.MethodPost)
}
