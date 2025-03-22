package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	graphhttphandler "github.com/msyamsula/portofolio/domain/graph/http"
	"github.com/msyamsula/portofolio/domain/telemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	appName := "grap-http"

	// load env
	godotenv.Load(".env")

	telemetry.InitializeTelemetryTracing(appName, os.Getenv("JAEGER_HOST"))
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("error in port format", err)
	}

	graphHandler := &graphhttphandler.Service{}

	// router
	apiPrefix := "/graph"
	r := mux.NewRouter()
	r.HandleFunc(fmt.Sprintf("%s%s", apiPrefix, "/{algo}"), http.HandlerFunc(graphHandler.InitGraph(http.HandlerFunc(graphHandler.Algorithm))))

	// Set up CORS options
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                              // Allow all origins (adjust for security)
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "HEAD"}, // Allow methods
		AllowedHeaders:   []string{"Content-Type"},                   // Allow headers
		AllowCredentials: true,                                       // Allows credentials (cookies, authorization headers)
	})

	corsHandler := c.Handler(r)

	// handler
	http.Handle("/", otelhttp.NewHandler(corsHandler, "")) // use otelhttp for telemetry
	http.Handle("/metrics", promhttp.Handler())

	// server
	err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
