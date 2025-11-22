package handler

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/msyamsula/portofolio/backend-app/message/persistence"
	"github.com/msyamsula/portofolio/backend-app/message/service"
)

func newSqsConsumer(c SqsConfig) *sqsConsumer {
	ctx := context.Background()

	// Load AWS config (uses env vars, shared creds, IAM role, etc.)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	return &sqsConsumer{
		client: sqs.NewFromConfig(cfg),
		url:    c.QueueUrl,
		svc:    c.Svc,
	}
}

type sqsConsumer struct {
	client *sqs.Client
	url    string

	svc service.Service
}

var (
	eventSend = "SEND"
)

func (s *sqsConsumer) Consume() {
	for {
		// Receive messages (long polling)
		ctx := context.Background()
		resp, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:              aws.String(s.url),
			MaxNumberOfMessages:   10,
			WaitTimeSeconds:       20, // long poll
			VisibilityTimeout:     30, // seconds
			MessageAttributeNames: []string{"All"},
		})

		if err != nil {
			log.Printf("receive error: %v", err)
			continue
		}

		if len(resp.Messages) == 0 {
			continue
		}

		for _, msg := range resp.Messages {

			m := persistence.Message{}
			err = m.UnmarshalJSON([]byte(*msg.Body))
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(m)

			err = s.process(m)
			if err != nil {
				continue // do not delete and keep it in queue for further processing
			}

			// Delete after successful processing
			_, err := s.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(s.url),
				ReceiptHandle: msg.ReceiptHandle,
			})
			if err != nil {
				log.Printf("delete error: %v", err)
			}
		}
	}
}

func (s *sqsConsumer) process(m persistence.Message) error {
	if m.Event == eventSend {
		_, err := s.svc.InsertMessage(context.Background(), m)
		return err
	}

	return nil
}
