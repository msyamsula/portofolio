package handler

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/msyamsula/portofolio/backend-app/message/persistence"
	"github.com/msyamsula/portofolio/backend-app/message/service"
)

type sqsConsumer struct {
	client *sqs.Client
	url    string

	svc service.Service
}

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
			fmt.Println("Received:", *msg.Body)

			// TODO: process message here
			// processMessage(msg.Body)
			// s.svc.InsertUnreadMessage()

			m := persistence.Message{}
			err = m.UnmarshalJSON([]byte(*msg.Body))
			if err != nil {
				fmt.Println(err)
			}
			// fmt.Println(aws.String(*msg.Body))
			fmt.Println(string(*msg.Body))
			fmt.Println(m)

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

func main() {

}
