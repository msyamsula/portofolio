package url

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/msyamsula/portofolio/database/postgres"
	"github.com/msyamsula/portofolio/database/redis"
	"github.com/msyamsula/portofolio/domain/hasher"
	"github.com/msyamsula/portofolio/domain/url/repository"
)

func TestIntegration(t *testing.T) {
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
		Word:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890,./[]!@£$%^&*()_+",
	})

	dep := Dependencies{
		Repo: repository.New(repository.Dependencies{
			Persistence: pg,
			Cache:       re,
		}),
		Hasher: ha,
	}

	svc := New(dep)
	fmt.Println(svc)

	ctx := context.Background()
	fmt.Println(svc.SetShortUrl(ctx, "http://mantap/2"))
	fmt.Println(svc.GetLongUrl(ctx, "wikwik"))

}
