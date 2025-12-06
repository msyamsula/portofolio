package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/backend-app/pkg/cache"
	"github.com/msyamsula/portofolio/backend-app/pkg/logger"
	"github.com/msyamsula/portofolio/backend-app/pkg/randomizer"
	"github.com/msyamsula/portofolio/backend-app/pkg/telemetry"
	"github.com/msyamsula/portofolio/backend-app/user/handler"
	pb "github.com/msyamsula/portofolio/backend-app/user/proto"
	"github.com/msyamsula/portofolio/backend-app/user/service"
	externaloauth "github.com/msyamsula/portofolio/backend-app/user/service/external-oauth"
	internaltoken "github.com/msyamsula/portofolio/backend-app/user/service/internal-token"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	appName = "user"
	env     = os.Getenv("ENVIRONMENT")

	tracerCollectorEndpoint = os.Getenv("TRACER_COLLECTOR_ENDPOINT")

	port     = os.Getenv("PORT")
	grpcPort = os.Getenv("GRPC_PORT")

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
		fmt.Println("TRACER_COLLECTOR_ENDPOINT:", tracerCollectorEndpoint)
		fmt.Println("PORT:", port)
		fmt.Println("AWS_ACCESS_KEY_ID:", awsAccessKeyId)
		fmt.Println("AWS_SECRET_ACCESS_KEY:", awsSecretAccessKey)
		fmt.Println("AWS_REGION:", awsRegion)
		fmt.Println("USER_LOGIN_TTL:", userLoginTtl)
		fmt.Println("JWT_TOKEN_TTL:", jwtTokenTtl)
		fmt.Println("GOOGLE_CLIENT_ID:", googleClientId)
		fmt.Println("GOOGLE_CLIENT_SECRET:", googleClientSecret)
		fmt.Println("GOOGLE_REDIRECT_URL:", googleRedirectUrl)
		fmt.Println("APP_TOKEN_SECRET:", appTokenSecret)
		fmt.Println("REDIS_HOST:", redisHost)
		fmt.Println("REDIS_PORT:", redisPort)
		fmt.Println("GRPC_PORT:", grpcPort)
	}
}

func route(r *mux.Router, h handler.Handler) *mux.Router {
	// url
	r.HandleFunc("/metrics", promhttp.Handler().ServeHTTP) // endpoint exporter, for prometheus scrapping
	r.HandleFunc("/login", h.GoogleRedirectUrl).Methods(http.MethodGet)
	r.HandleFunc("/google/callback", h.GetAppTokenForGoogle).Methods(http.MethodGet)
	return r
}

func initHandler() *handler.CombineHandler {
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

	internalToken := internaltoken.NewInternalToken(internaltoken.InternalTokenConfig{
		AppTokenSecret: appTokenSecret,
		AppTokenTtl:    time.Duration(tokenTtl * int64(time.Minute)),
	})

	sessionManager := cache.NewRedis(cache.RedisConfig{
		Host: redisHost,
		Port: redisPort,
		Env:  env,
	}, &redis.Options{
		MaxRetries:      10,
		MinRetryBackoff: 5,
		MaxRetryBackoff: 10,
		ReadTimeout:     1 * time.Second,
		WriteTimeout:    1 * time.Second,
		MinIdleConns:    3,
		MaxIdleConns:    5,
		MaxActiveConns:  10,
	})

	externalOauth := externaloauth.NewAuthService(externaloauth.AuthConfig{
		GoogleOauthConfig: &oauth2.Config{
			ClientID:     googleClientId,
			RedirectURL:  googleRedirectUrl,
			ClientSecret: googleClientSecret,
			Endpoint:     google.Endpoint,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		},
	})

	h := handler.New(handler.Config{
		Svc: service.NewService(service.ServiceConfig{
			External:          externalOauth,
			Internal:          internalToken,
			SessionManagement: sessionManager,
			UserLoginTtl:      time.Duration(userTtl * int64(time.Hour)),
		}),
		Randomizer: randomizer.NewStringRandomizer(randomizer.StringRandomizerConfig{
			Size:          20,
			CharacterPool: "abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		}),
		InternalToken: internalToken,
	})

	return h
}

func main() {

	telemetry.InitializeTelemetryTracing(appName, tracerCollectorEndpoint)
	logger.InitLogger()

	h := initHandler()

	// create server routes
	r := mux.NewRouter()
	r = route(r, h)

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
	httpAddress := fmt.Sprintf("0.0.0.0:%s", port)
	server := &http.Server{
		Addr:    httpAddress,
		Handler: finalHandler,
	}

	go func() {
		logger.Logger.Infof("http server starting at %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatalf("server failed: %v", err)
		}
	}()

	go func() {

		logger.Logger.Infof("gRPC server listening at %s", grpcPort)
		grpcAddress := fmt.Sprintf(":%s", grpcPort)
		lis, err := net.Listen("tcp", grpcAddress)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		s := grpc.NewServer()
		pb.RegisterExampleServiceServer(s, h)

		reflection.Register(s)

		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
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
