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
	"github.com/msyamsula/portofolio/backend-app/url-shortener/cache"
	"github.com/msyamsula/portofolio/backend-app/url-shortener/handler"
	"github.com/msyamsula/portofolio/backend-app/url-shortener/persistent"
	"github.com/msyamsula/portofolio/backend-app/url-shortener/services"
	"github.com/msyamsula/portofolio/binary/telemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	pgPassword = os.Getenv("POSTGRES_PASSWORD")
	pgUsername = os.Getenv("POSTGRES_USER")
	pgDbName   = os.Getenv("POSTGRES_DB")
	pgHost     = os.Getenv("POSTGRES_HOST")
	pgPort     = os.Getenv("POSTGRES_PORT")

	redisHost     = os.Getenv("REDIS_HOST")
	redisPort     = os.Getenv("REDIS_PORT")
	redisPassword = os.Getenv("REDIS_PASSWORD")

	jaegerHost = os.Getenv("JAEGER_HOST")

	characterPool = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyz"

	callbackUri = os.Getenv("CALLBACK_URI")
)

func Route(r *mux.Router) *mux.Router {
	appName := "url"

	// initialize instrumentation
	telemetry.InitializeTelemetryTracing(appName, jaegerHost)

	h := handler.New(handler.Config{
		Svc: services.New(services.Config{
			Persistence: persistent.New(persistent.Config{
				Username: pgUsername,
				Name:     pgDbName,
				Password: pgPassword,
				Host:     pgHost,
				Port:     pgPort,
			}),
			// Cache: nil,
			Cache: cache.New(cache.Config{
				Host:     redisHost,
				Port:     redisPort,
				Password: redisPassword,
				Ttl:      5 * time.Minute,
			}),
			CharacterPool: characterPool,
			Size:          10,
			CallbackUri:   callbackUri,
		}),
	})

	// url
	r.HandleFunc("/short", h.Short).Methods(http.MethodGet)
	r.HandleFunc("/{shortUrl}", h.Redirect).Methods(http.MethodGet)
	return r
}

func main() {

	// create server routes
	r := mux.NewRouter()
	r = Route(r)

	http.Handle("/metrics", promhttp.Handler()) // endpoint exporter, for prometheus scrapping

	// cors option
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                              // Allow all origins (adjust for security)
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "HEAD"}, // Allow methods
		AllowedHeaders:   []string{"Content-Type"},                   // Allow headers
		AllowCredentials: true,                                       // Allows credentials (cookies, authorization headers)
	})
	tracedHandler := otelhttp.NewHandler(r, "")
	finalHandler := cors.Handler(tracedHandler)

	// server start
	port := 10000
	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: finalHandler,
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
