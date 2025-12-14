package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/backend-app/friend/handler"
	repo "github.com/msyamsula/portofolio/backend-app/friend/persistent"
	"github.com/msyamsula/portofolio/backend-app/friend/service"
	"github.com/msyamsula/portofolio/backend-app/pkg/telemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	appName = "friend"
	env     = os.Getenv("ENVIRONMENT")

	pgPassword = os.Getenv("POSTGRES_PASSWORD")
	pgUsername = os.Getenv("POSTGRES_USER")
	pgDbName   = os.Getenv("POSTGRES_DB")
	pgHost     = os.Getenv("POSTGRES_HOST")
	pgPort     = os.Getenv("POSTGRES_PORT")

	tracerCollectorEndpoint = os.Getenv("TRACER_COLLECTOR_ENDPOINT")

	port = os.Getenv("PORT")
)

func init() {
	if env != "production" {
		log.Printf("ENVIRONMENT: %s", env)
		log.Printf("POSTGRES_USER: %s", pgUsername)
		log.Printf("POSTGRES_PASSWORD: %s", pgPassword)
		log.Printf("POSTGRES_DB: %s", pgDbName)
		log.Printf("POSTGRES_HOST: %s", pgHost)
		log.Printf("POSTGRES_PORT: %s", pgPort)
		log.Printf("TRACER_COLLECTOR_ENDPOINT: %s", tracerCollectorEndpoint)
		log.Printf("PORT: %s", port)
	}
}

func createLogFile() *os.File {
	// Include file name and line number in log output
	log.SetFlags(log.LstdFlags | log.Llongfile)

	// Open (or create) a log file
	if env != "production" {
		log.Println("local")
		f, err := os.OpenFile(fmt.Sprintf("%s_log", appName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("failed to open log file: %v\n", err)
			return nil
		}

		multiOuput := io.MultiWriter(os.Stdout, f)
		log.SetOutput(multiOuput)

		return f
	}

	return nil

}

func route(r *mux.Router) *mux.Router {

	// var h handler.Handler
	h := handler.New(handler.Config{
		Svc: service.New(service.ServiceConfig{
			Persistent: repo.NewPostgres(repo.PostgresConfig{
				Username: pgUsername,
				Password: pgPassword,
				DbName:   pgDbName,
				Host:     pgHost,
				Port:     pgPort,
			}),
		}),
	})

	// url
	r.HandleFunc("/friends", h.AddFriend).Methods(http.MethodPost)
	r.HandleFunc("/friends", h.GetFriends).Methods(http.MethodGet)
	return r
}

func main() {

	// initialize instrumentation
	flush := telemetry.InitializeTelemetryTracing(appName, tracerCollectorEndpoint)
	defer flush()

	f := createLogFile()
	defer f.Close()

	// create server routes
	r := mux.NewRouter()
	r = route(r)

	r.HandleFunc("/metrics", promhttp.Handler().ServeHTTP) // endpoint exporter, for prometheus scrapping
	tracedHandler := otelhttp.NewHandler(r, "friend http server")

	// cors option
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                              // Allow all origins (adjust for security)
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "HEAD"}, // Allow methods
		AllowedHeaders:   []string{"Content-Type"},                   // Allow headers
		AllowCredentials: true,                                       // Allows credentials (cookies, authorization headers)
	})
	finalHandler := cors.Handler(tracedHandler)

	// server start
	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", port),
		Handler: finalHandler,
	}

	log.Println("server starting...")
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
