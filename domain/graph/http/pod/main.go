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
	// load env
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("error in loading env", err)
	}

	appName := "grap-http"
	telemetry.InitializeTelemetryTracing(appName, os.Getenv("JAEGER_HOST"))
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("error in port format", err)
	}

	// build the service dependencies
	// ------------------------------
	graphHandler := graphhttphandler.Service{}

	apiPrefix := "/api/graph"
	r := mux.NewRouter()

	preflight := func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, HEAD")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				return
			}

			next.ServeHTTP(w, r)
		}
	}

	// API listing
	r.HandleFunc(fmt.Sprintf("%s%s", apiPrefix, "/{algo}"), preflight(http.HandlerFunc(graphHandler.InitGraph(http.HandlerFunc(graphHandler.Algorithm)))))

	// Set up CORS options
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                              // Allow all origins (adjust for security)
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "HEAD"}, // Allow methods
		AllowedHeaders:   []string{"Content-Type"},                   // Allow headers
		AllowCredentials: true,                                       // Allows credentials (cookies, authorization headers)
	})

	r.Use(c.Handler)

	otelHandler := otelhttp.NewHandler(r, "")
	http.Handle("/", otelHandler) // use otelhttp for telemetry
	http.Handle("/metrics", promhttp.Handler())

	err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
