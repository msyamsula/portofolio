package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	ch := make(chan os.Signal, 1)

	producer := &producer{
		awsRegion: os.Getenv("AWS_REGION"),
		topicArn:  os.Getenv("SNS_TOPIC_ARN"),
	}
	consumer := &consumer{}
	go consumer.Start(ch)
	go func() {
		time.Sleep(2 * time.Second)
		producer.Publish(context.Background(), fmt.Sprintf("Hello at %s", time.Now().Format(time.RFC3339)))
	}()

	// Wait for interrupt signal
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch

	// Graceful shutdown
	if err := consumer.Stop(); err != nil {
		log.Printf("server forced to shutdown: %v", err)
		return
	}

	log.Println("Server stopped gracefully")
}
