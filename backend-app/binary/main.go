package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	friendHandler "github.com/msyamsula/portofolio/backend-app/domain/friend/handler"
	friendRepo "github.com/msyamsula/portofolio/backend-app/domain/friend/repository"
	friendSvc "github.com/msyamsula/portofolio/backend-app/domain/friend/service"
	graphHandler "github.com/msyamsula/portofolio/backend-app/domain/graph/handler"
	graphSvc "github.com/msyamsula/portofolio/backend-app/domain/graph/service"
	healthcheckHandler "github.com/msyamsula/portofolio/backend-app/domain/healthcheck/handler"
	healthcheckSvc "github.com/msyamsula/portofolio/backend-app/domain/healthcheck/service"
	messageHandler "github.com/msyamsula/portofolio/backend-app/domain/message/handler"
	messageRepo "github.com/msyamsula/portofolio/backend-app/domain/message/repository"
	messageSvc "github.com/msyamsula/portofolio/backend-app/domain/message/service"
	urlShortenerHandler "github.com/msyamsula/portofolio/backend-app/domain/url-shortener/handler"
	urlShortenerRepo "github.com/msyamsula/portofolio/backend-app/domain/url-shortener/repository"
	urlShortenerSvc "github.com/msyamsula/portofolio/backend-app/domain/url-shortener/service"
	userHandler "github.com/msyamsula/portofolio/backend-app/domain/user/handler"
	userIntegration "github.com/msyamsula/portofolio/backend-app/domain/user/integration"
	userSvc "github.com/msyamsula/portofolio/backend-app/domain/user/service"
	infraDB "github.com/msyamsula/portofolio/backend-app/infrastructure/database/postgres"
	redisInf "github.com/msyamsula/portofolio/backend-app/infrastructure/database/redis"
	infraHttp "github.com/msyamsula/portofolio/backend-app/infrastructure/http/middleware"
	infraLogger "github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
	infraMetrics "github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/metrics"
	infraSpan "github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/span"
)

// Config holds application configuration
type Config struct {
	// Server configuration
	ServerPort string

	// Base URL
	BaseURL string

	// PostgreSQL configuration
	PostgresHost string
	PostgresPort string
	PostgresUser string
	PostgresPass string
	PostgresDB   string
	PostgresSSL  string

	// Redis configuration
	RedisHost     string
	RedisPassword string
	RedisDB       int

	// Telemetry configuration
	TelemetryCollectorEndpoint string
	ServiceName                string
	MetricsPushInterval        time.Duration

	// Logger configuration
	LogLevel  string
	LogFormat string

	// OAuth configuration
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// Token configuration
	AppTokenSecret string
	AppTokenTTL    time.Duration
}

func main() {
	// Load configuration
	cfg := loadConfig()

	// Initialize logger (must be first for proper logging)
	if err := initLogger(cfg); err != nil {
		infraLogger.Error("failed to initialize logger", map[string]any{"error": err})
		os.Exit(1)
	}
	defer func() {
		if err := infraLogger.Shutdown(context.Background()); err != nil {
			infraLogger.ErrorError("failed to shutdown logger", err, nil)
		}
	}()

	// Initialize telemetry
	spanClient := initTelemetry(cfg)
	if spanClient != nil {
		defer func() {
			if err := spanClient.Shutdown(context.Background()); err != nil {
				infraLogger.ErrorError("failed to shutdown telemetry", err, nil)
			}
		}()
	}

	metricsClient, instruments := initMetrics(cfg)
	if metricsClient != nil {
		defer func() {
			if err := metricsClient.Shutdown(context.Background()); err != nil {
				infraLogger.ErrorError("failed to shutdown metrics", err, nil)
			}
		}()
	}

	// Initialize database connections
	db := initPostgres(cfg)
	rdb := initRedis(cfg)

	// Initialize domains and register routes
	router := setupRouter(cfg, db, rdb, instruments)

	// Start server
	startServer(router, cfg.ServerPort)
}

func setupRouter(cfg Config, db *sqlx.DB, rdb *redis.Client, instruments *infraMetrics.Instruments) *mux.Router {
	r := mux.NewRouter()

	// Apply common middleware to all routes
	r.Use(infraHttp.MetricsMiddleware(instruments))
	r.Use(infraHttp.TracingMiddleware(cfg.ServiceName))

	// Initialize URL Shortener domain
	urlShortenerRepo := urlShortenerRepo.NewRepository(db, rdb)
	urlShortenerSvc := urlShortenerSvc.New(cfg.BaseURL, urlShortenerRepo)
	urlShortenerHandler := urlShortenerHandler.New(urlShortenerSvc)

	// Initialize Graph domain
	graphSvc := graphSvc.New()
	graphHandler := graphHandler.New(graphSvc)

	// Initialize Friend domain
	friendRepo := friendRepo.NewPostgresRepository(db)
	friendSvc := friendSvc.New(friendRepo)
	friendHandler := friendHandler.New(friendSvc)

	// Initialize Message domain
	messageRepo := messageRepo.NewPostgresRepository(db)
	messageSvc := messageSvc.New(messageRepo)
	messageHandler := messageHandler.New(messageSvc)

	// Initialize User domain
	googleAuthService := userIntegration.NewGoogleAuthService(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)
	userSvc := userSvc.New(googleAuthService, cfg.AppTokenSecret, cfg.AppTokenTTL)
	userHandler := userHandler.New(userSvc)

	// Initialize Healthcheck domain
	healthcheckSvc := healthcheckSvc.New()
	healthcheckHandler := healthcheckHandler.New(healthcheckSvc)

	// Register URL Shortener routes
	urlShortenerChain := infraHttp.Chain(
		infraHttp.ContentTypeMiddleware,
		infraHttp.RecoveryMiddleware,
		infraHttp.LoggingMiddleware,
		infraHttp.CORSMiddleware,
		infraHttp.TracingMiddleware("url-shortener"),
	)
	urlShortenerRouter := r.PathPrefix("/url").Subrouter()
	urlShortenerRouter.Use(urlShortenerChain)
	urlShortenerHandler.RegisterRoutes(urlShortenerRouter)

	// Register Graph routes
	graphChain := infraHttp.Chain(
		infraHttp.ContentTypeMiddleware,
		infraHttp.RecoveryMiddleware,
		infraHttp.LoggingMiddleware,
		infraHttp.CORSMiddleware,
		infraHttp.TracingMiddleware("graph"),
	)
	graphRouter := r.PathPrefix("/graph").Subrouter()
	graphRouter.Use(graphChain)
	graphHandler.RegisterRoutes(graphRouter)

	// Register Friend routes
	friendChain := infraHttp.Chain(
		infraHttp.ContentTypeMiddleware,
		infraHttp.RecoveryMiddleware,
		infraHttp.LoggingMiddleware,
		infraHttp.CORSMiddleware,
		infraHttp.TracingMiddleware("friend"),
	)
	friendRouter := r.PathPrefix("/friend").Subrouter()
	friendRouter.Use(friendChain)
	friendHandler.RegisterRoutes(friendRouter)

	// Register Message routes
	messageChain := infraHttp.Chain(
		infraHttp.ContentTypeMiddleware,
		infraHttp.RecoveryMiddleware,
		infraHttp.LoggingMiddleware,
		infraHttp.CORSMiddleware,
		infraHttp.TracingMiddleware("message"),
	)
	messageRouter := r.PathPrefix("/message").Subrouter()
	messageRouter.Use(messageChain)
	messageHandler.RegisterRoutes(messageRouter)

	// Register User routes
	userChain := infraHttp.Chain(
		infraHttp.ContentTypeMiddleware,
		infraHttp.RecoveryMiddleware,
		infraHttp.LoggingMiddleware,
		infraHttp.CORSMiddleware,
		infraHttp.TracingMiddleware("user"),
	)
	userRouter := r.PathPrefix("/user").Subrouter()
	userRouter.Use(userChain)
	userHandler.RegisterRoutes(userRouter)

	// Register Healthcheck routes
	healthcheckChain := infraHttp.Chain(
		infraHttp.ResponseTimeMiddleware,
		infraHttp.TracingMiddleware("healthcheck"),
	)
	healthcheckRouter := r.PathPrefix("/health").Subrouter()
	healthcheckRouter.Use(healthcheckChain)
	healthcheckHandler.RegisterRoutes(healthcheckRouter)

	// Register Swagger UI route
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	return r
}

func startServer(router *mux.Router, port string) {
	addr := ":" + port

	infraLogger.Info("starting server", map[string]any{
		"address": addr,
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			infraLogger.ErrorError("server error", err, nil)
		}
	}()

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	// Note: importing "os/signal" would create the variable
	// signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	infraLogger.Info("shutting down server", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		infraLogger.ErrorError("server shutdown error", err, nil)
	}
	infraLogger.Info("server stopped", nil)
}

func loadConfig() Config {
	return Config{
		ServerPort:                 getEnv("SERVER_PORT", "5000"),
		BaseURL:                    getEnv("BASE_URL", "https://short.est"),
		PostgresHost:               getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:               getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:               getEnv("POSTGRES_USER", "postgres"),
		PostgresPass:               getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:                 getEnv("POSTGRES_DB", "portofolio"),
		PostgresSSL:                getEnv("POSTGRES_SSL", "disable"),
		RedisHost:                  getEnv("REDIS_HOST", "localhost:6379"),
		RedisPassword:              getEnv("REDIS_PASSWORD", ""),
		RedisDB:                    getEnvInt("REDIS_DB", "0"),
		TelemetryCollectorEndpoint: getEnv("OTEL_COLLECTOR_ENDPOINT", "localhost:4317"),
		ServiceName:                getEnv("SERVICE_NAME", "backend-app"),
		MetricsPushInterval:        getDurationEnv("OTEL_METRICS_INTERVAL", 15*time.Second),
		LogLevel:                   getEnv("LOG_LEVEL", "INFO"),
		LogFormat:                  getEnv("LOG_FORMAT", "TEXT"),
		GoogleClientID:             getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:         getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:          getEnv("GOOGLE_REDIRECT_URL", "http://localhost:5000/user/google/callback"),
		AppTokenSecret:             getEnv("APP_TOKEN_SECRET", generateRandomSecret()),
		AppTokenTTL:                getDurationEnv("APP_TOKEN_TTL", 24*time.Hour),
	}
}

func initLogger(cfg Config) error {
	ctx := context.Background()

	// Parse log level
	var level infraLogger.Level
	switch cfg.LogLevel {
	case "DEBUG":
		level = infraLogger.DebugLevel
	case "INFO":
		level = infraLogger.InfoLevel
	case "WARN":
		level = infraLogger.WarnLevel
	case "ERROR":
		level = infraLogger.ErrorLevel
	default:
		level = infraLogger.InfoLevel
	}

	// Parse log format
	var format infraLogger.Format
	if cfg.LogFormat == "JSON" {
		format = infraLogger.JSONFormat
	} else {
		format = infraLogger.TextFormat
	}

	// Initialize logger with dual output (stdout + OTLP)
	if err := infraLogger.Init(ctx, infraLogger.Config{
		ServiceName:       cfg.ServiceName,
		CollectorEndpoint: cfg.TelemetryCollectorEndpoint,
		Insecure:          true,
		Environment:       getEnv("ENVIRONMENT", "local"),
		LogsEnabled:       true,
		Level:             level,
		Format:            format,
		TimeFormat:        time.RFC3339,
	}); err != nil {
		return err
	}

	infraLogger.Info("logger initialized", map[string]any{
		"level":        cfg.LogLevel,
		"format":       cfg.LogFormat,
		"otlp_enabled": true,
	})

	return nil
}

func initTelemetry(cfg Config) *infraSpan.Client {
	ctx := context.Background()
	spanClient, err := infraSpan.NewClient(ctx, infraSpan.Config{
		ServiceName:       cfg.ServiceName,
		CollectorEndpoint: cfg.TelemetryCollectorEndpoint,
		Insecure:          true,
		SampleRate:        1.0,
		Environment:       getEnv("ENVIRONMENT", "local"),
	})
	if err != nil {
		infraLogger.WarnError("failed to create telemetry client", err, map[string]any{"service": cfg.ServiceName})
		return nil
	}
	return spanClient
}

func initMetrics(cfg Config) (*infraMetrics.Client, *infraMetrics.Instruments) {
	ctx := context.Background()
	metricsClient, err := infraMetrics.NewClient(ctx, infraMetrics.Config{
		ServiceName:       cfg.ServiceName,
		CollectorEndpoint: cfg.TelemetryCollectorEndpoint,
		Insecure:          true,
		PushInterval:      cfg.MetricsPushInterval,
		Environment:       getEnv("ENVIRONMENT", "local"),
	})
	if err != nil {
		infraLogger.WarnError("failed to create metrics client", err, map[string]any{"service": cfg.ServiceName})
		return nil, nil
	}

	instruments, err := infraMetrics.NewInstruments(metricsClient.Meter(cfg.ServiceName))
	if err != nil {
		infraLogger.WarnError("failed to create metrics instruments", err, nil)
		return metricsClient, nil
	}

	return metricsClient, instruments
}

func initPostgres(cfg Config) *sqlx.DB {
	ctx := context.Background()
	db, err := infraDB.NewPostgresClient(ctx, infraDB.Config{
		User:     cfg.PostgresUser,
		Host:     cfg.PostgresHost,
		Password: cfg.PostgresPass,
		Database: cfg.PostgresDB,
		Port:     cfg.PostgresPort,
		SSLMode:  cfg.PostgresSSL,
	})
	if err != nil {
		infraLogger.ErrorError("failed to connect to postgres", err, map[string]any{
			"host":     cfg.PostgresHost,
			"port":     cfg.PostgresPort,
			"database": cfg.PostgresDB,
		})
		os.Exit(1)
	}
	infraLogger.Info("connected to postgres", map[string]any{
		"host":     cfg.PostgresHost,
		"port":     cfg.PostgresPort,
		"database": cfg.PostgresDB,
	})
	return db
}

func initRedis(cfg Config) *redis.Client {
	ctx := context.Background()
	rdb := redisInf.NewRedisClient(ctx, redisInf.RedisConfig{
		Host:     cfg.RedisHost,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := redisInf.PingRedis(ctx, rdb); err != nil {
		infraLogger.ErrorError("failed to connect to redis", err, map[string]any{
			"host": cfg.RedisHost,
			"db":   cfg.RedisDB,
		})
		os.Exit(1)
	}
	infraLogger.Info("connected to redis", map[string]any{
		"host": cfg.RedisHost,
		"db":   cfg.RedisDB,
	})
	return rdb
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue string) int {
	valueStr := getEnv(key, defaultValue)
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return 0
	}
	return int(value)
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, defaultValue.String())
	parsed, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func generateRandomSecret() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
