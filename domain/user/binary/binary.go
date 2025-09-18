package binary

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/msyamsula/portofolio/binary/postgres"
	"github.com/msyamsula/portofolio/binary/redis"
	"github.com/msyamsula/portofolio/binary/telemetry"
	userhttp "github.com/msyamsula/portofolio/domain/user/http"
	useroauth "github.com/msyamsula/portofolio/domain/user/oauth"
	"github.com/msyamsula/portofolio/domain/user/repository"
	"github.com/msyamsula/portofolio/domain/user/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func initUserHandler(userSvc *service.Service) *userhttp.Handler {

	handler := userhttp.New(userhttp.Dependencies{
		Service: userSvc,
	})
	return handler

}

func initDataLayer() (*postgres.Postgres, *redis.Redis) {

	pg := postgres.New(postgres.Config{
		Username: os.Getenv("POSTGRES_PASSWORD"),
		Password: os.Getenv("POSTGRES_USERNAME"),
		DbName:   os.Getenv("POSTGRES_DB"),
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
	})

	ttl, err := strconv.Atoi(os.Getenv("REDIS_TTL"))
	if err != nil {
		log.Fatal("redis ttl error")
	}
	redisTtl := time.Duration(ttl) * time.Millisecond
	re := redis.New(redis.Config{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		Ttl:      redisTtl,
	})

	return pg, re
}

func initGoogleSigninService(userSvc *service.Service) *useroauth.Service {
	return useroauth.New(useroauth.Dependencies{
		GoogleClientId:      os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleRedirectOauth: os.Getenv("GOOGLE_REDIRECT_OAUTH"),
		GoogleSecret:        os.Getenv("GOOGLE_SECRET"),
		UserSvc:             userSvc,
		RedirectChat:        os.Getenv("REDIRECT_CHAT"),
		OauthStateLength:    25,
		OauthCharacters:     os.Getenv("HASHER_CHARACTER_POOL"),
	})
}

func Run() {
	appName := "user"

	// load env
	godotenv.Load(".env")

	// initialize instrumentation
	telemetry.InitializeTelemetryTracing(appName, os.Getenv("JAEGER_HOST"))

	pg, re := initDataLayer()

	userSvc := service.New(service.Dependencies{
		Persistence: &repository.Persistence{
			Postgres: pg,
		},
		Cache: &repository.Cache{
			Redis: re,
		},
	})

	// create userHandler
	userHandler := initUserHandler(userSvc)
	googleUserOauthHandler := initGoogleSigninService(userSvc)

	// create server routes
	r := mux.NewRouter()

	r.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good"))
	})

	// google sign in
	r.HandleFunc("/access/token", googleUserOauthHandler.HandleCallback)
	r.HandleFunc("/google/signin", googleUserOauthHandler.HandleLogin)
	// user
	r.HandleFunc("/user/friend", userHandler.ManageFriend)
	r.HandleFunc("/user", userHandler.ManageUser)

	// cors option
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                              // Allow all origins (adjust for security)
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "HEAD"}, // Allow methods
		AllowedHeaders:   []string{"Content-Type"},                   // Allow headers
		AllowCredentials: true,                                       // Allows credentials (cookies, authorization headers)
	})
	corsHandler := c.Handler(r)

	// server handler
	http.Handle("/", otelhttp.NewHandler(corsHandler, "")) // use otelhttp for telemetry
	http.Handle("/metrics", promhttp.Handler())            // endpoint exporter, for prometheus scrapping

	// server start
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("error in port format", err)
	}
	err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
