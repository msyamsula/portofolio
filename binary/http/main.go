package main

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
	"github.com/msyamsula/portofolio/domain/google"
	graphhttp "github.com/msyamsula/portofolio/domain/graph/http"
	messagehttp "github.com/msyamsula/portofolio/domain/message/http"
	messagerepo "github.com/msyamsula/portofolio/domain/message/repository"
	messagesvc "github.com/msyamsula/portofolio/domain/message/service"
	"github.com/msyamsula/portofolio/domain/url/hasher"
	urlhttp "github.com/msyamsula/portofolio/domain/url/http"
	urlrepo "github.com/msyamsula/portofolio/domain/url/repository"
	url "github.com/msyamsula/portofolio/domain/url/service"
	userhttp "github.com/msyamsula/portofolio/domain/user/http"
	"github.com/msyamsula/portofolio/domain/user/repository"
	"github.com/msyamsula/portofolio/domain/user/service"
	"github.com/prometheus/client_golang/prometheus"
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

func initUrlHandler(pg *postgres.Postgres, re *redis.Redis) *urlhttp.Handler {

	hasherLength, err := strconv.Atoi(os.Getenv("HASHER_LENGTH"))
	if err != nil {
		fmt.Println("error in parsing hash length")
		log.Fatal(err)
	}
	ha := hasher.New(hasher.Config{
		Length: int64(hasherLength),
		Word:   os.Getenv("HASHER_CHARACTER_POOL"),
	})

	dep := urlhttp.Dependencies{
		UrlService: url.New(url.Dependencies{
			Repo: urlrepo.New(urlrepo.Dependencies{
				Persistence: pg,
				Cache:       re,
			}),
			Host:   os.Getenv("HASHER_HOST"),
			Hasher: ha,
		}),
	}
	urlHandler := urlhttp.New(dep)
	return urlHandler

}

func initGraphHandler() *graphhttp.Service {
	return &graphhttp.Service{}
}

func initMessageHandler(pg *postgres.Postgres) *messagehttp.Handler {

	handler := messagehttp.New(messagehttp.Dependencies{
		Service: messagesvc.New(messagesvc.Dependencies{
			Persistence: messagerepo.New(pg),
		}),
	})
	return handler

}

func initGoogleSigninService(userSvc *service.Service) *google.Service {
	return google.New(google.Dependencies{
		GoogleClientId:      os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleRedirectOauth: os.Getenv("GOOGLE_REDIRECT_OAUTH"),
		GoogleSecret:        os.Getenv("GOOGLE_SECRET"),
		UserSvc:             userSvc,
		RedirectChat:        os.Getenv("REDIRECT_CHAT"),
		OauthStateLength:    25,
		OauthCharacters:     os.Getenv("HASHER_CHARACTER_POOL"),
	})
}

func main() {
	appName := "backend"

	// load env
	godotenv.Load(".env")

	// initialize instrumentation
	telemetry.InitializeTelemetryTracing(appName, os.Getenv("JAEGER_HOST"))

	// register prometheus metrics
	prometheus.MustRegister(urlhttp.HashCounter)
	prometheus.MustRegister(urlhttp.RedirectCounter)

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
	urlHandler := initUrlHandler(pg, re)
	graphHandler := initGraphHandler()
	messageHandler := initMessageHandler(pg)
	googleSigninHandler := initGoogleSigninService(userSvc)

	// create server routes
	r := mux.NewRouter()
	// google sign in
	r.HandleFunc("/access/token", googleSigninHandler.HandleCallback)
	r.HandleFunc("/google/signin", googleSigninHandler.HandleLogin)
	// message
	r.HandleFunc("/message", messageHandler.ManageMesage)
	// user
	r.HandleFunc("/user/friend", userHandler.ManageFriend)
	r.HandleFunc("/user", userHandler.ManageUser)
	// graph
	r.HandleFunc("/graph/{algo}", http.HandlerFunc(graphHandler.InitGraph(http.HandlerFunc(graphHandler.Algorithm))))
	// url
	r.HandleFunc("/short", urlHandler.HashUrl)
	r.HandleFunc("/{shortUrl}", urlHandler.RedirectShortUrl)

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
	http.Handle("/metrics", promhttp.Handler())

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
