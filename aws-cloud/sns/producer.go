package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type producer struct {
	awsRegion string
	topicArn  string
}

func (p *producer) Publish(ctx context.Context, message string) error {

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("failed to load AWS config:", err)
	}

	// Create SNS client
	client := sns.NewFromConfig(cfg)

	// Publish a message
	input := &sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: &p.topicArn,
	}

	result, err := client.Publish(context.TODO(), input)
	if err != nil {
		log.Fatal("failed to publish message:", err)
		return err
	}

	log.Println("Message published, ID:", *result.MessageId)
	return nil
}
