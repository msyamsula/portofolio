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
	"github.com/msyamsula/portofolio/backend-app/observability/logger"
	"github.com/msyamsula/portofolio/backend-app/observability/telemetry"
	"github.com/msyamsula/portofolio/backend-app/url-shortener/cache"
	"github.com/msyamsula/portofolio/backend-app/url-shortener/handler"
	"github.com/msyamsula/portofolio/backend-app/url-shortener/persistent"
	"github.com/msyamsula/portofolio/backend-app/url-shortener/services"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
)

var (
	appName = "url-shortener"
	env     = os.Getenv("ENVIRONMENT")

	pgPassword = os.Getenv("POSTGRES_PASSWORD")
	pgUsername = os.Getenv("POSTGRES_USER")
	pgDbName   = os.Getenv("POSTGRES_DB")
	pgHost     = os.Getenv("POSTGRES_HOST")
	pgPort     = os.Getenv("POSTGRES_PORT")

	redisHost = os.Getenv("REDIS_HOST")
	redisPort = os.Getenv("REDIS_PORT")

	tracerCollectorEndpoint = os.Getenv("TRACER_COLLECTOR_ENDPOINT")

	characterPool = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyz"

	callbackUri = os.Getenv("CALLBACK_URI")
	port        = os.Getenv("PORT")

	awsAccessKeyId     = os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsRegion          = os.Getenv("AWS_REGION")

	dynamoTable = os.Getenv("DYNAMO_TABLE")
)

func printEnv() {
	if env != "production" {
		logger.Logger.Info("ENVIRONMENT:", env)
		logger.Logger.Info("POSTGRES_PASSWORD:", pgPassword)
		logger.Logger.Info("POSTGRES_USER:", pgUsername)
		logger.Logger.Info("POSTGRES_DB:", pgDbName)
		logger.Logger.Info("POSTGRES_HOST:", pgHost)
		logger.Logger.Info("POSTGRES_PORT:", pgPort)
		logger.Logger.Info("REDIS_HOST:", redisHost)
		logger.Logger.Info("REDIS_PORT:", redisPort)
		logger.Logger.Info("TRACER_COLLECTOR_ENDPOINT:", tracerCollectorEndpoint)
		logger.Logger.Info("CALLBACK_URI:", callbackUri)
		logger.Logger.Info("PORT:", port)
		logger.Logger.Info("AWS_ACCESS_KEY_ID:", awsAccessKeyId)
		logger.Logger.Info("AWS_SECRET_ACCESS_KEY:", awsSecretAccessKey)
		logger.Logger.Info("AWS_REGION:", awsRegion)
		logger.Logger.Info("DYNAMO_TABLE:", dynamoTable)
	}
}

func route(r *mux.Router) *mux.Router {

	// var h handler.Handler
	h := handler.New(handler.Config{
		Svc: services.New(services.Config{
			Persistence: persistent.NewPostgres(persistent.PostgresConfig{
				Username: pgUsername,
				Name:     pgDbName,
				Password: pgPassword,
				Host:     pgHost,
				Port:     pgPort,
				Attributes: []attribute.KeyValue{
					{
						Key:   "app",
						Value: attribute.StringValue(appName),
					},
				},
				Env: env,
			}),
			Cache: cache.NewRedis(cache.RedisConfig{
				Host: redisHost,
				Port: redisPort,
				Ttl:  5 * time.Minute,
				Env:  env,
			}),
			CharacterPool: characterPool,
			Size:          10,
			CallbackUri:   callbackUri,
		}),
	})

	// url
	r.HandleFunc("/metrics", promhttp.Handler().ServeHTTP) // endpoint exporter, for prometheus scrapping
	r.HandleFunc("/short", h.Short).Methods(http.MethodGet)
	r.HandleFunc("/{shortUrl}", h.Redirect).Methods(http.MethodGet)
	return r
}

func main() {

	printEnv()
	logger.InitLogger()

	// create server routes
	r := mux.NewRouter()
	r = route(r)

	tracedHandler := otelhttp.NewHandler(r, "http server")

	// initialize instrumentation
	shutdown := telemetry.InitializeTelemetryTracing(appName, tracerCollectorEndpoint)
	defer shutdown()

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

	logger.Logger.Info("server starting")
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
		logger.Logger.Panicf("server forced to shutdown: %v", err)
	}

	logger.Logger.Info("Server stopped gracefully")
}
