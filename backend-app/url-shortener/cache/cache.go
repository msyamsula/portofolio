package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	configPkg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	redisPkg "github.com/redis/go-redis/v9"
)

type Repository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
}

func New(cfg Config) Repository {
	client := redisPkg.NewClient(&redisPkg.Options{
		Addr:           fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:       cfg.Password, // No password set
		DB:             0,            // Use default DB
		Protocol:       2,            // Connection protocol
		PoolSize:       10,
		TLSConfig:      &tls.Config{},
		DialTimeout:    2 * time.Second,
		PoolTimeout:    1 * time.Second,
		MaxActiveConns: 7,
		MinIdleConns:   3,
		MaxIdleConns:   3,
		ReadTimeout:    0,
		WriteTimeout:   0,
	})
	return &redis{
		db:  client,
		ttl: cfg.Ttl,
	}
}

/*
you need to have this in your env runtime

	export AWS_ACCESS_KEY_ID="your-access-key"
	export AWS_SECRET_ACCESS_KEY="your-secret-key"
	export AWS_SESSION_TOKEN="optional-temporary-token"
*/
func NewDynamo(config DynamoConfig) Repository {
	// Load default AWS config (from ~/.aws/credentials, env variables, or IAM role)
	cfg, err := configPkg.LoadDefaultConfig(context.TODO(),
		configPkg.WithRegion(config.Region), // change to your region

	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create DynamoDB client
	svc := dynamodb.NewFromConfig(cfg)

	fmt.Println("connected to dynamo db")

	return &dynamo{
		ttl:       config.Ttl,
		tablename: config.TableName,
		region:    config.Region,
		conn:      svc,
	}
}
