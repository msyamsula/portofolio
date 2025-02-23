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
	"github.com/msyamsula/portofolio/domain/hasher"
	urlhttp "github.com/msyamsula/portofolio/domain/url/handler/http"
	"github.com/msyamsula/portofolio/domain/url/repository"
	url "github.com/msyamsula/portofolio/domain/url/service"
)

func main() {
	// load env
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("error in loading env", err)
	}

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

	r.HandleFunc(fmt.Sprintf("%s%s", apiPrefix, "/short"), urlSevice.GetShortUrl)
	r.HandleFunc(fmt.Sprintf("%s%s", apiPrefix, "/redirect/{shortUrl}"), urlSevice.RedirectShortUrl)

	http.Handle("/", r)

	err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
