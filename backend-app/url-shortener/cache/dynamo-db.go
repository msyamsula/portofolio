package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type dynamo struct {
	ttl       time.Duration
	tablename string
	region    string

	conn *dynamodb.Client
}

func (r *dynamo) Get(ctx context.Context, key string) (string, error) {

	limit := int32(1)
	dynamoQuery := &dynamodb.QueryInput{
		TableName:              aws.String(r.tablename),
		KeyConditionExpression: aws.String("short_url = :short_url"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":short_url": &types.AttributeValueMemberS{Value: key},
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
		return "", nil
	}

	attr := resp.Items[0]["long_url"]
	if attr == nil {
		return "", errors.New("long url attribute not found")
	}

	stringAttr, ok := attr.(*types.AttributeValueMemberS)
	if !ok {
		return "", errors.New("long url is not string")
	}

	return string(stringAttr.Value), nil
}

func (r *dynamo) Set(ctx context.Context, key, value string) error {

	ttl := time.Now().Unix() + int64(r.ttl.Seconds()) // ttl is in unix for dynamodb
	item := map[string]types.AttributeValue{
		"short_url": &types.AttributeValueMemberS{Value: key},
		"long_url":  &types.AttributeValueMemberS{Value: value},
		"ttl":       &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", ttl)}, // optional TTL
	}
	_, err := r.conn.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &r.tablename,
		Item:      item,
	})
	return err
}
