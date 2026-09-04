package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
)

func work(message kafka.Message) error {
	var payload messagePayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return err
	}

	fmt.Println(payload.ID)
	return errors.New("work failed")
}

func add(message kafka.Message) error {
	var payload messagePayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return err
	}

	fmt.Println(payload.ID + 1)
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	printConsumer := NewConsumer(bootstrapServer, topic, groupID, work)
	// addConsumer := NewConsumer(bootstrapServer, topic, "syamsul-add-consumer-group", add)

	errors := make(chan error, 2)
	go func() { errors <- printConsumer.Consume(ctx) }()
	// go func() { errors <- addConsumer.Consume(ctx) }()

	if err := <-errors; err != nil {
		log.Fatal(err)
	}
}
