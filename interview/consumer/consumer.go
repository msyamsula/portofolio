package main

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

const (
	bootstrapServer = "localhost:9092"
	topic           = "syamsul"
	groupID         = "syamsul-consumer-group"
)

type messagePayload struct {
	ID int `json:"id"`
}

type Handler func(kafka.Message) error

type Consumer struct {
	server  string
	topic   string
	groupID string
	handler Handler
}

func NewConsumer(server, topic, groupID string, handler Handler) Consumer {
	return Consumer{
		server:  server,
		topic:   topic,
		groupID: groupID,
		handler: handler,
	}
}

func (c Consumer) Consume(ctx context.Context) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{c.server},
		Topic:          c.topic,
		GroupID:        c.groupID,
		CommitInterval: 0,
		StartOffset:    kafka.FirstOffset,
	})
	defer reader.Close()

	log.Printf("consuming topic %q as group %q", c.topic, c.groupID)
	for {
		message, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		if err := c.handler(message); err != nil {
			log.Printf("processing message at offset %d: %v", message.Offset, err)
			continue
		}
		if err := reader.CommitMessages(ctx, message); err != nil {
			log.Printf("committing message at offset %d: %v", message.Offset, err)
		}
	}
}
