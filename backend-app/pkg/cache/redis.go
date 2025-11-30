package cache

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/msyamsula/portofolio/backend-app/observability/logger"
	"github.com/redis/go-redis/extra/redisotel/v9"
	redisPkg "github.com/redis/go-redis/v9"
)

type redis struct {
	db *redisPkg.Client
}

func NewRedis(cfg RedisConfig, options *redisPkg.Options) *redisPkg.Client {
	var tlsConfig *tls.Config
	if cfg.Env == "production" {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	options.Addr = fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	options.TLSConfig = tlsConfig

	client := redisPkg.NewClient(options)
	ping := client.Ping(context.Background())
	_, err := ping.Result()
	if err != nil {
		logger.Logger.Fatalf("ping %s failed, %v", options.Addr, err.Error())
	}
	redisotel.InstrumentTracing(client)
	return client
}
