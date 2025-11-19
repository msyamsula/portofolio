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
