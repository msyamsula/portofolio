package main

import (
	"fmt"
	"log"
	"net"
	"time"

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
	port := 8000
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	pg := postgres.New(postgres.Config{
		Username: "admin",
		Password: "admin",
		DbName:   "postgres",
		Host:     "127.0.0.1",
		Port:     "5432",
	})

	re := redis.New(redis.Config{
		Host:     "127.0.0.1",
		Port:     "6379",
		Password: "admin",
		Ttl:      300 * time.Second,
	})

	ha := hasher.New(hasher.Config{
		Length: 10,
		Word:   "abcdefghijklmnopqrstuvwxyz1234567890!@£$%^&*()_+",
		Host:   "http://syamsul.com",
	})

	dep := urlgrpc.Dependencies{
		UrlService: url.New(url.Dependencies{
			Repo: repository.New(repository.Dependencies{
				Persistence: pg,
				Cache:       re,
			}),
			Hasher: ha,
		}),
	}

	pb.RegisterUrlShortenerServer(grpcServer, urlgrpc.New(dep))
	reflection.Register(grpcServer)
	grpcServer.Serve(lis)
}
