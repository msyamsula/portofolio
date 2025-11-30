package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/backend-app/pkg/cache"
	"github.com/msyamsula/portofolio/backend-app/pkg/logger"
	"github.com/msyamsula/portofolio/backend-app/pkg/telemetry"
	"github.com/msyamsula/portofolio/backend-app/user/handler"
	"github.com/msyamsula/portofolio/backend-app/user/service"
	externaloauth "github.com/msyamsula/portofolio/backend-app/user/service/external-oauth"
	internaltoken "github.com/msyamsula/portofolio/backend-app/user/service/internal-token"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	appName = "user"
	env     = os.Getenv("ENVIRONMENT")

	pgPassword = os.Getenv("POSTGRES_PASSWORD")
	pgUsername = os.Getenv("POSTGRES_USER")
	pgDbName   = os.Getenv("POSTGRES_DB")
	pgHost     = os.Getenv("POSTGRES_HOST")
	pgPort     = os.Getenv("POSTGRES_PORT")

	jaegerHost = os.Getenv("TRACER_COLLECTOR_ENDPOINT")

	port = os.Getenv("PORT")

	awsAccessKeyId     = os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsRegion          = os.Getenv("AWS_REGION")

	userLoginTtl = os.Getenv("USER_LOGIN_TTL")
	jwtTokenTtl  = os.Getenv("JWT_TOKEN_TTL")

	googleClientId     = os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	googleRedirectUrl  = os.Getenv("GOOGLE_REDIRECT_URL")

	appTokenSecret = os.Getenv("APP_TOKEN_SECRET")

	redisHost = os.Getenv("REDIS_HOST")
	redisPort = os.Getenv("REDIS_PORT")
)

func init() {
	if env != "production" {
		fmt.Println("ENVIRONMENT:", env)
		fmt.Println("POSTGRES_USER:", pgUsername)
		fmt.Println("POSTGRES_DB:", pgDbName)
		fmt.Println("POSTGRES_HOST:", pgHost)
		fmt.Println("POSTGRES_PORT:", pgPort)
		fmt.Println("TRACER_COLLECTOR_ENDPOINT:", jaegerHost)
		fmt.Println("PORT:", port)
		fmt.Println("AWS_REGION:", awsRegion)
		fmt.Println("USER_LOGIN_TTL:", userLoginTtl)
		fmt.Println("JWT_TOKEN_TTL:", jwtTokenTtl)
	}
}

func route(r *mux.Router) *mux.Router {

	// initialize instrumentation
	telemetry.InitializeTelemetryTracing(appName, jaegerHost)

	// var h handler.Handler
	var err error
	var userTtl int64
	userTtl, err = strconv.ParseInt(userLoginTtl, 10, 64)
	if err != nil {
		logger.Logger.Panic("invalid user ttl")
	}

	var tokenTtl int64
	tokenTtl, err = strconv.ParseInt(jwtTokenTtl, 10, 64)
	if err != nil {
		logger.Logger.Panic("invalid jwt token ttl")
	}

	h := handler.New(handler.Config{
		Svc: service.NewService(service.ServiceConfig{
			External: externaloauth.NewAuthService(
				externaloauth.AuthConfig{
					GoogleOauthConfig: &oauth2.Config{
						ClientID:     googleClientId,
						ClientSecret: googleClientSecret,
						RedirectURL:  googleRedirectUrl,
						Endpoint:     google.Endpoint,
						Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
					},
				},
			),
			Internal: internaltoken.NewInternalToken(
				internaltoken.InternalTokenConfig{
					AppTokenSecret: appTokenSecret,
					AppTokenTtl:    time.Duration(tokenTtl * int64(time.Hour)),
				},
			),
			SessionManagement: cache.NewRedis(
				cache.RedisConfig{
					Host: redisHost,
					Port: redisPort,
					Env:  env,
				},
				&redis.Options{
					// retries and connection pool
					MaxRetries:     20,
					DialTimeout:    10 * time.Second,
					ReadTimeout:    1 * time.Second,
					WriteTimeout:   1 * time.Second,
					PoolTimeout:    20 * time.Second,
					MinIdleConns:   5,
					MaxIdleConns:   10,
					MaxActiveConns: 10,
				},
			),
			UserLoginTtl: time.Duration(userTtl * int64(time.Hour)),
		}),
	})

	// url
	r.HandleFunc("/metrics", promhttp.Handler().ServeHTTP) // endpoint exporter, for prometheus scrapping
	r.HandleFunc("/login", h.GoogleRedirectUrl).Methods(http.MethodGet)
	r.HandleFunc("/callback", h.GetAppTokenForGoogle).Methods(http.MethodGet)
	return r
}

func main() {

	telemetry.InitializeTelemetryTracing(appName, jaegerHost)
	logger.InitLogger()

	// create server routes
	r := mux.NewRouter()
	r = route(r)

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

	logger.Logger.Info("server starting...")
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatalf("server failed: %v", err)
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
