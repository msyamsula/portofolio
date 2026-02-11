package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	healthcheckHandler "github.com/msyamsula/portofolio/backend-app/domain/healthcheck/handler"
	healthcheckService "github.com/msyamsula/portofolio/backend-app/domain/healthcheck/service"
	urlShortenerHandler "github.com/msyamsula/portofolio/backend-app/domain/url-shortener/handler"
	"github.com/msyamsula/portofolio/backend-app/domain/url-shortener/repository"
	urlShortenerService "github.com/msyamsula/portofolio/backend-app/domain/url-shortener/service"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/database/postgres"
	redisinf "github.com/msyamsula/portofolio/backend-app/infrastructure/database/redis"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/http/middleware"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/span"
)

// Config holds application configuration
type Config struct {
	ServerPort    string
	BaseURL       string
	PostgresHost  string
	PostgresPort  string
	PostgresUser  string
	PostgresPass  string
	PostgresDB    string
	PostgresSSL   string
	RedisHost     string
	RedisPassword string
	RedisDB       int
	// Telemetry configuration
	TelemetryCollectorEndpoint string
	ServiceName                string
	// Logger configuration
	LogLevel string
	LogFormat string
}

func main() {
	// Load configuration
	cfg := loadConfig()

	// Initialize logger (must be first for proper logging)
	loggerClient := initLogger(cfg)
	if loggerClient != nil {
		defer func() {
			if err := loggerClient.Shutdown(context.Background()); err != nil {
				loggerClient.Logger.ErrorError("failed to shutdown logger", err, nil)
			}
		}()
	}

	// Initialize telemetry
	spanClient := initTelemetry(cfg, loggerClient.Logger)
	if spanClient != nil {
		defer func() {
			if err := spanClient.Shutdown(context.Background()); err != nil {
				loggerClient.Logger.ErrorError("failed to shutdown telemetry", err, nil)
			}
		}()
	}

	// Initialize dependencies
	db := initPostgres(cfg, loggerClient.Logger)
	rdb := initRedis(cfg, loggerClient.Logger)

	// Initialize URL shortener domain
	urlShortenerRepo := repository.NewRepository(db, rdb)
	urlShortenerSvc := urlShortenerService.New(cfg.BaseURL, urlShortenerRepo)
	urlHandler := urlShortenerHandler.New(urlShortenerSvc)

	// Initialize healthcheck domain
	healthcheckSvc := healthcheckService.New()
	healthHandler := healthcheckHandler.New(healthcheckSvc)

	// Setup HTTP server
	domainHandler := domainHandler{
		url:         urlHandler,
		healthcheck: healthHandler,
	}
	router := setupServer(domainHandler)

	// Start server
	startServer(router, cfg.ServerPort, loggerClient.Logger)
}

func loadConfig() Config {
	return Config{
		ServerPort:                 getEnv("SERVER_PORT", "5000"),
		BaseURL:                    getEnv("BASE_URL", "https://short.est"),
		PostgresHost:               getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:               getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:               getEnv("POSTGRES_USER", "postgres"),
		PostgresPass:               getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:                 getEnv("POSTGRES_DB", "urlshortener"),
		PostgresSSL:                getEnv("POSTGRES_SSLMODE", "disable"),
		RedisHost:                  getEnv("REDIS_HOST", "localhost:6379"),
		RedisPassword:              getEnv("REDIS_PASSWORD", ""),
		RedisDB:                    0,
		TelemetryCollectorEndpoint: getEnv("OTEL_COLLECTOR_ENDPOINT", "localhost:4317"),
		ServiceName:                getEnv("SERVICE_NAME", "url-shortener"),
		LogLevel:                   getEnv("LOG_LEVEL", "INFO"),
		LogFormat:                  getEnv("LOG_FORMAT", "TEXT"),
	}
}

func initLogger(cfg Config) *logger.Client {
	ctx := context.Background()

	// Parse log level
	var level logger.Level
	switch cfg.LogLevel {
	case "DEBUG":
		level = logger.DebugLevel
	case "INFO":
		level = logger.InfoLevel
	case "WARN":
		level = logger.WarnLevel
	case "ERROR":
		level = logger.ErrorLevel
	default:
		level = logger.InfoLevel
	}

	// Parse log format
	var format logger.Format
	if cfg.LogFormat == "JSON" {
		format = logger.JSONFormat
	} else {
		format = logger.TextFormat
	}

	// Initialize logger client with dual output (stdout + OTLP)
	loggerClient, err := logger.NewClient(ctx, logger.Config{
		ServiceName:       cfg.ServiceName,
		CollectorEndpoint: cfg.TelemetryCollectorEndpoint,
		Insecure:          true,
		Environment:       getEnv("ENVIRONMENT", "local"),
		LogsEnabled:       true, // Enable OTLP export
		Level:             level,
		Format:            format,
		TimeFormat:        time.RFC3339,
	})
	if err != nil {
		// Fallback to default logger if initialization fails
		defaultLogger := logger.Default()
		defaultLogger.ErrorError("failed to initialize logger client, using default", err, nil)
		return &logger.Client{Logger: defaultLogger}
	}

	loggerClient.Logger.Info("logger initialized", map[string]any{
		"level": cfg.LogLevel,
		"format": cfg.LogFormat,
		"otlp_enabled": true,
	})

	return loggerClient
}

func initTelemetry(cfg Config, log *logger.Logger) *span.Client {
	ctx := context.Background()

	// Initialize telemetry span client using infrastructure
	spanClient, err := span.NewClient(ctx, span.Config{
		ServiceName:       cfg.ServiceName,
		CollectorEndpoint: cfg.TelemetryCollectorEndpoint,
		Insecure:          true,
		SampleRate:        1.0, // 100% sampling for development
		Environment:       getEnv("ENVIRONMENT", "local"),
	})
	if err != nil {
		log.WarnError("failed to create telemetry client", err, map[string]any{"service": cfg.ServiceName})
		return nil
	}

	// Set global propagator for trace context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Info("telemetry initialized", map[string]any{"service": cfg.ServiceName})
	return spanClient
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initPostgres(cfg Config, log *logger.Logger) *sqlx.DB {
	ctx := context.Background()
	db, err := postgres.NewPostgresClient(ctx, postgres.Config{
		Host:     cfg.PostgresHost,
		Port:     cfg.PostgresPort,
		User:     cfg.PostgresUser,
		Password: cfg.PostgresPass,
		Database: cfg.PostgresDB,
		SSLMode:  cfg.PostgresSSL,
	})
	if err != nil {
		log.ErrorError("failed to connect to postgres", err, map[string]any{
			"host": cfg.PostgresHost,
			"port": cfg.PostgresPort,
			"database": cfg.PostgresDB,
		})
		os.Exit(1)
	}
	log.Info("connected to postgres", map[string]any{
		"host": cfg.PostgresHost,
		"port": cfg.PostgresPort,
		"database": cfg.PostgresDB,
	})
	return db
}

func initRedis(cfg Config, log *logger.Logger) *redis.Client {
	ctx := context.Background()
	rdb := redisinf.NewRedisClient(ctx, redisinf.RedisConfig{
		Host:     cfg.RedisHost,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := redisinf.PingRedis(ctx, rdb); err != nil {
		log.ErrorError("failed to connect to redis", err, map[string]any{
			"host": cfg.RedisHost,
			"db": cfg.RedisDB,
		})
		os.Exit(1)
	}
	log.Info("connected to redis", map[string]any{
		"host": cfg.RedisHost,
		"db": cfg.RedisDB,
	})
	return rdb
}

type domainHandler struct {
	url         *urlShortenerHandler.Handler
	healthcheck *healthcheckHandler.Handler
}

func setupServer(h domainHandler) *mux.Router {
	r := mux.NewRouter()

	// Health check - tracing only (no auth/cors for k8s probes)
	healthRouter := r.PathPrefix("/health").Subrouter()
	healthRouter.Use(middleware.TracingMiddleware("healthcheck"))
	healthRouter.HandleFunc("", h.healthcheck.Check).Methods("GET")

	// Common middleware for all URL shortener routes
	urlShortenerChain := middleware.Chain(
		middleware.TracingMiddleware("urlshortener"),
		middleware.RecoveryMiddleware,
		middleware.LoggingMiddleware,
		middleware.CORSMiddleware,
		xPortofolioMiddleware,
	)

	// URL shortener admin routes - Using /url prefix
	urlShortenerRouter := r.PathPrefix("/url").Subrouter()
	urlShortenerRouter.Use(urlShortenerChain)
	urlShortenerRouter.HandleFunc("/shorten", h.url.Shorten).Methods("POST")

	// Short code redirect - Must be last to catch unmatched paths
	redirectRouter := r.PathPrefix("/").Subrouter()
	redirectRouter.Use(urlShortenerChain)
	redirectRouter.HandleFunc("/{shortCode}", h.url.Redirect).Methods("GET")

	return r
}

func xPortofolioMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for x-portofolio header
		if r.Header.Get("x-portofolio") == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"missing x-portofolio header","meta":{"responseTime":0}}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func startServer(router *mux.Router, port string, log *logger.Logger) {
	addr := ":" + port
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("starting server", map[string]any{"address": addr})
		log.Info("available endpoints", nil)
		log.Info("  GET  /health", nil)
		log.Info("  GET  /{shortCode}           (x-portofolio header required)", nil)
		log.Info("  POST /shorten                (x-portofolio header required)", nil)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.ErrorError("server error", err, nil)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.ErrorError("server shutdown error", err, nil)
	}
	log.Info("server stopped", nil)
}
