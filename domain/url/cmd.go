package url

import (
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
	urlhttp "github.com/msyamsula/portofolio/domain/url/http"
	urlrepo "github.com/msyamsula/portofolio/domain/url/repository"
	url "github.com/msyamsula/portofolio/domain/url/service"
	"github.com/prometheus/client_golang/prometheus"
)

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

	repo := urlrepo.New(urlrepo.Dependencies{
		Persistence: pg,
		Cache:       re,
	})

	service := url.New(url.Dependencies{
		Repo:          repo,
		Host:          os.Getenv("HASHER_HOST"),
		Length:        int64(hasherLength),
		CharacterPool: os.Getenv("HASHER_CHARACTER_POOL"),
	})

	urlHandler := urlhttp.New(urlhttp.Dependencies{
		UrlService: service,
	})
	return urlHandler

}

func Run(r *mux.Router) {
	appName := "url"

	// load env
	godotenv.Load(".env")

	// initialize instrumentation
	telemetry.InitializeTelemetryTracing(appName, os.Getenv("JAEGER_HOST"))

	// register prometheus metrics
	prometheus.MustRegister(urlhttp.HashCounter)
	prometheus.MustRegister(urlhttp.RedirectCounter)

	pg, re := initDataLayer()

	urlHandler := initUrlHandler(pg, re)

	// url
	r.HandleFunc("/short", urlHandler.HashUrl).Methods(http.MethodGet)
	r.HandleFunc("/{shortUrl}", urlHandler.RedirectShortUrl).Methods(http.MethodGet)

}
