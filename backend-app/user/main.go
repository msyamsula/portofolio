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
	"github.com/msyamsula/portofolio/backend-app/user/handler"
	"github.com/msyamsula/portofolio/backend-app/user/persistent"
	"github.com/msyamsula/portofolio/backend-app/user/service"
	"github.com/msyamsula/portofolio/telemetryv2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	appName = "user"
	env     = os.Getenv("ENVIRONMENT")

	pgPassword = os.Getenv("POSTGRES_PASSWORD")
	pgUsername = os.Getenv("POSTGRES_USER")
	pgDbName   = os.Getenv("POSTGRES_DB")
	pgHost     = os.Getenv("POSTGRES_HOST")
	pgPort     = os.Getenv("POSTGRES_PORT")

	jaegerHost = os.Getenv("JAEGER_HOST")

	port = os.Getenv("PORT")

	awsAccessKeyId     = os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsRegion          = os.Getenv("AWS_REGION")
)

func init() {
	if env != "production" {
		fmt.Println("ENVIRONMENT:", env)
		fmt.Println("POSTGRES_USER:", pgUsername)
		fmt.Println("POSTGRES_DB:", pgDbName)
		fmt.Println("POSTGRES_HOST:", pgHost)
		fmt.Println("POSTGRES_PORT:", pgPort)
		fmt.Println("JAEGER_HOST:", jaegerHost)
		fmt.Println("PORT:", port)
		fmt.Println("AWS_REGION:", awsRegion)
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

	// initialize instrumentation
	telemetryv2.InitializeTelemetryTracing(appName, jaegerHost)

	// var h handler.Handler
	h := handler.New(handler.Config{
		Svc: service.New(service.ServiceConfig{
			Persistence: persistent.NewPostgres(persistent.PostgresConfig{
				Username: pgUsername,
				Password: pgPassword,
				DbName:   pgDbName,
				Host:     pgHost,
				Port:     port,
			}),
		}),
	})

	// url
	r.HandleFunc("/signup", h.InsertUser).Methods(http.MethodGet)
	r.HandleFunc("/login", h.GetUser).Methods(http.MethodGet)
	return r
}

func main() {

	f := createLogFile()
	defer f.Close()

	// create server routes
	r := mux.NewRouter()
	r = route(r)

	r.HandleFunc("/metrics", promhttp.Handler().ServeHTTP) // endpoint exporter, for prometheus scrapping
	tracedHandler := otelhttp.NewHandler(r, "")

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
