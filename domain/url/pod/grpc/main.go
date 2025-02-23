package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/msyamsula/portofolio/database/postgres"
	"github.com/msyamsula/portofolio/database/redis"
	"github.com/msyamsula/portofolio/domain/hasher"
	urlgrpc "github.com/msyamsula/portofolio/domain/url/handler/grpc"
	"github.com/msyamsula/portofolio/domain/url/handler/grpc/pb"
	"github.com/msyamsula/portofolio/domain/url/repository"
	url "github.com/msyamsula/portofolio/domain/url/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	// load env
	err := godotenv.Load("domain/url/pod/grpc/.env")
	if err != nil {
		log.Fatal("error in loading env", err)
	}

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

	dep := urlgrpc.Dependencies{
		UrlService: url.New(url.Dependencies{
			Repo: repository.New(repository.Dependencies{
				Persistence: pg,
				Cache:       re,
			}),
			Host:   os.Getenv("HASHER_HOST"),
			Hasher: ha,
		}),
	}

	// server part
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	pb.RegisterUrlShortenerServer(grpcServer, urlgrpc.New(dep))
	reflection.Register(grpcServer)
	log.Println("starting server on port", port)
	grpcServer.Serve(lis)
}
