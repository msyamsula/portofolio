package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	AWS_REGION            = os.Getenv("AWS_REGION")
	SNS_TOPIC_ARN         = os.Getenv("SNS_TOPIC_ARN")
	AWS_ACCESS_KEY_ID     = os.Getenv("AWS_ACCESS_KEY_ID")
	AWS_SECRET_ACCESS_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	ch := make(chan os.Signal, 1)

	log.Println("AWS_REGION:", AWS_REGION)
	log.Println("SNS_TOPIC_ARN:", SNS_TOPIC_ARN)
	log.Println("AWS_ACCESS_KEY_ID:", AWS_ACCESS_KEY_ID)
	log.Println("AWS_SECRET_ACCESS_KEY:", AWS_SECRET_ACCESS_KEY)

	producer := &producer{
		awsRegion: AWS_REGION,
		topicArn:  SNS_TOPIC_ARN,
	}
	consumer := &consumer{}
	go consumer.Start(ch)
	go producer.foreverloop()

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
