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
	"github.com/msyamsula/portofolio/domain/telemetry"
	"github.com/msyamsula/portofolio/domain/url/hasher"
	urlhttp "github.com/msyamsula/portofolio/domain/url/http"
	"github.com/msyamsula/portofolio/domain/url/repository"
	url "github.com/msyamsula/portofolio/domain/url/service"
	"github.com/msyamsula/portofolio/tech-stack/postgres"
	"github.com/msyamsula/portofolio/tech-stack/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func initUrlHandler() *urlhttp.Handler {

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

	hasherLength, err := strconv.Atoi(os.Getenv("HASHER_LENGTH"))
	ha := hasher.New(hasher.Config{
		Length: int64(hasherLength),
		Word:   os.Getenv("HASHER_CHARACTER_POOL"),
	})

	dep := urlhttp.Dependencies{
		UrlService: url.New(url.Dependencies{
			Repo: repository.New(repository.Dependencies{
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

func main() {
	appName := "url-http"

	// load env
	godotenv.Load(".env")

	// initialize instrumentation
	telemetry.InitializeTelemetryTracing(appName, os.Getenv("JAEGER_HOST"))

	// register prometheus metrics
	prometheus.MustRegister(urlhttp.HashCounter)
	prometheus.MustRegister(urlhttp.RedirectCounter)

	// create handler
	urlHandler := initUrlHandler()

	// create server routes
	r := mux.NewRouter()
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
