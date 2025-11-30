package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	configPkg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type dynamo struct {
	ttl       time.Duration
	tablename string
	region    string

	conn *dynamodb.Client
}

/*
you need to have this in your env runtime

	export AWS_ACCESS_KEY_ID="your-access-key"
	export AWS_SECRET_ACCESS_KEY="your-secret-key"
	export AWS_SESSION_TOKEN="optional-temporary-token"
*/
func NewDynamo(config DynamoConfig) *dynamo {
	// Load default AWS config (from ~/.aws/credentials, env variables, or IAM role)
	cfg, err := configPkg.LoadDefaultConfig(context.TODO(),
		configPkg.WithRegion(config.Region), // change to your region

	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create DynamoDB client
	svc := dynamodb.NewFromConfig(cfg)

	log.Println("connected to dynamo db")

	return &dynamo{
		ttl:       config.Ttl,
		tablename: config.TableName,
		region:    config.Region,
		conn:      svc,
	}
}

func (r *dynamo) Get(ctx context.Context, key string) (string, error) {

	limit := int32(1)
	dynamoQuery := &dynamodb.QueryInput{
		TableName:              aws.String(r.tablename),
		KeyConditionExpression: aws.String("identifier = :identifier"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":identifier": &types.AttributeValueMemberS{Value: key},
		},
		Limit: &limit,
	}
	resp, err := r.conn.Query(ctx, dynamoQuery)
	if err != nil {
		log.Printf("failed to get item: %v\n", err)
		return "", err
	}

	if len(resp.Items) == 0 {
		log.Println("Item not found")
		return "", errors.New("item not found")
	}

	attr := resp.Items[0]["value"]
	if attr == nil {
		return "", errors.New("value attribute not found")
	}

	stringAttr, ok := attr.(*types.AttributeValueMemberS)
	if !ok {
		return "", errors.New("value is not string")
	}

	return string(stringAttr.Value), nil
}

func (r *dynamo) Set(ctx context.Context, key, value string) error {

	ttl := time.Now().Unix() + int64(r.ttl.Seconds()) // ttl is in unix for dynamodb
	item := map[string]types.AttributeValue{
		"identifier": &types.AttributeValueMemberS{Value: key},
		"value":      &types.AttributeValueMemberS{Value: value},
		"ttl":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", ttl)}, // optional TTL
	}
	_, err := r.conn.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &r.tablename,
		Item:      item,
	})
	return err
}
