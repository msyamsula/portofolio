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
	"github.com/msyamsula/portofolio/database/postgres"
	"github.com/msyamsula/portofolio/database/redis"
	"github.com/msyamsula/portofolio/domain/telemetry"
	"github.com/msyamsula/portofolio/domain/url/hasher"
	urlhttp "github.com/msyamsula/portofolio/domain/url/http"
	"github.com/msyamsula/portofolio/domain/url/repository"
	url "github.com/msyamsula/portofolio/domain/url/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	// load env
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("error in loading env", err)
	}

	appName := "url-http"
	telemetry.InitializeTelemetryTracing(appName, os.Getenv("JAEGER_HOST"), os.Getenv("PROMOTEHUS_HOST"))

	// build the service dependencies
	// ------------------------------
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("error in port format", err)
	}

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

	urlSevice := urlhttp.New(dep)

	apiPrefix := "/api/url"
	r := mux.NewRouter()

	// prometheus metrics
	hashCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hash_counter",
		Help: "number of shortener request",
	})
	prometheus.MustRegister(hashCounter)

	redirectCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "redirect_counter",
		Help: "number of redirect request",
	})
	prometheus.MustRegister(redirectCounter)

	// middleware
	hashCounterMiddleWare := func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r.URL.Path)
			hashCounter.Inc()
			next.ServeHTTP(w, r)
		}
	}

	redirectCounterMiddleWare := func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r.URL.Path)
			redirectCounter.Inc()
			next.ServeHTTP(w, r)
		}
	}

	// API listing
	r.HandleFunc(fmt.Sprintf("%s%s", apiPrefix, "/short"), hashCounterMiddleWare(http.HandlerFunc(urlSevice.HashUrl)))
	r.HandleFunc(fmt.Sprintf("%s%s", apiPrefix, "/redirect/{shortUrl}"), redirectCounterMiddleWare(http.HandlerFunc(urlSevice.RedirectShortUrl)))

	http.Handle("/", otelhttp.NewHandler(r, "")) // use otelhttp for telemetry
	http.Handle("/metrics", promhttp.Handler())

	err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
