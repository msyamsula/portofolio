//go:build integtaion

package grpc

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/msyamsula/portofolio/domain/url/grpc/pb"
	"github.com/msyamsula/portofolio/domain/url/hasher"
	"github.com/msyamsula/portofolio/domain/url/repository"
	url "github.com/msyamsula/portofolio/domain/url/service"
	"github.com/msyamsula/portofolio/tech-stack/postgres"
	"github.com/msyamsula/portofolio/tech-stack/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func TestIntegration(t *testing.T) {
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
	})

	dep := Dependencies{
		UrlService: url.New(url.Dependencies{
			Repo: repository.New(repository.Dependencies{
				Persistence: pg,
				Cache:       re,
			}),
			Hasher: ha,
		}),
	}
	pb.RegisterUrlShortenerServer(grpcServer, New(dep))
	reflection.Register(grpcServer)
	grpcServer.Serve(lis)
}
