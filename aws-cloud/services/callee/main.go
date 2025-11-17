package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Response struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func pinghandler(w http.ResponseWriter, r *http.Request) {
	resp := Response{}

	podName := os.Getenv("HOSTNAME")
	rn, _ := rand.Int(rand.Reader, big.NewInt(int64(1e7)))
	randomNumber := rn.Int64()

	resp.Code = http.StatusOK
	resp.Message = fmt.Sprintf("Hello from %s, with random: %d", podName, randomNumber)
	json.NewEncoder(w).Encode(resp)
}

func main() {

	// create server routes
	r := mux.NewRouter()

	// cors option
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                              // Allow all origins (adjust for security)
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "HEAD"}, // Allow methods
		AllowedHeaders:   []string{"Content-Type"},                   // Allow headers
		AllowCredentials: true,                                       // Allows credentials (cookies, authorization headers)
	})
	corsHandler := c.Handler(r)

	r.HandleFunc("/ping", pinghandler).Methods(http.MethodGet)

	// dummy server
	http.Handle("/", otelhttp.NewHandler(corsHandler, "")) // use otelhttp for telemetry

	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", 8000), nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
