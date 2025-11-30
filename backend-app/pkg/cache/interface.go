package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	configPkg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/msyamsula/portofolio/backend-app/pkg/logger"
	"github.com/redis/go-redis/extra/redisotel/v9"
	redisPkg "github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(c context.Context, key string) (string, error)
	Set(c context.Context, key string, value string, ttl time.Duration) error
	Del(c context.Context, key string) error
}

func NewRedis(cfg RedisConfig, options *redisPkg.Options) Cache {
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
	return &redis{
		db: client,
	}
}

/*
you need to have this in your env runtime

	export AWS_ACCESS_KEY_ID="your-access-key"
	export AWS_SECRET_ACCESS_KEY="your-secret-key"
	export AWS_SESSION_TOKEN="optional-temporary-token"
*/
func NewDynamo(config DynamoConfig) Cache {
	// Load default AWS config (from ~/.aws/credentials, env variables, or IAM role)
	cfg, err := configPkg.LoadDefaultConfig(context.TODO(),
		configPkg.WithRegion(config.Region), // change to your region

	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create DynamoDB client
	svc := dynamodb.NewFromConfig(cfg)

	logger.Logger.Info("connected to dynamo db")

	return &dynamo{
		tablename: config.TableName,
		region:    config.Region,
		conn:      svc,
	}
}
