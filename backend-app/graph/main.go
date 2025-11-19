package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/backend-app/graph/handler"
	"github.com/msyamsula/portofolio/backend-app/graph/service"
	"github.com/msyamsula/portofolio/telemetryv2"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	appName    = "graph"
	jaegerHost = os.Getenv("JAEGER_HOST")
	port       = os.Getenv("PORT")
)

func Route(r *mux.Router) *mux.Router {

	// initialize instrumentation
	telemetryv2.InitializeTelemetryTracing(appName, jaegerHost)
	h := handler.NewHandler(handler.Config{
		Service: service.New(),
	})

	r.HandleFunc("/graph/{algo}", h.Solve).Methods(http.MethodGet)

	return r

}

func main() {

	// create server routes
	r := mux.NewRouter()
	r = Route(r)

	tracedHandler := otelhttp.NewHandler(r, "")

	// cors option
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                              // Allow all origins (adjust for security)
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "HEAD"}, // Allow methods
		AllowedHeaders:   []string{"Content-Type"},                   // Allow headers
		AllowCredentials: true,                                       // Allows credentials (cookies, authorization headers)
	})
	corsHandler := cors.Handler(tracedHandler)

	port := 10000
	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: corsHandler,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}
