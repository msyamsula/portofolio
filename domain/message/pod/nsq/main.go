package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/msyamsula/portofolio/domain/message"
	"github.com/msyamsula/portofolio/domain/telemetry"
	"github.com/nsqio/go-nsq"
)

func main() {
	appName := "message-consumer"

	// load env
	godotenv.Load(".env")

	telemetry.InitializeTelemetryTracing(appName, os.Getenv("JAEGER_HOST"))
	var err error

	consumers := []message.Consumer{}
	lookupds := strings.Split(os.Getenv("LOOKUPDS"), ",")

	// start append consumers here
	{
		consumerSaveMessage, err := message.New(message.Config{
			Name:      message.ConfigSaveMessage,
			Lookupds:  lookupds,
			NsqConfig: nsq.NewConfig(),
		}, new(message.SaveMessageHandler))
		if err != nil {
			log.Fatal(err)
		}
		consumers = append(consumers, *consumerSaveMessage)
	}

	// consumer start
	for _, c := range consumers {
		err = c.Start()
		if err != nil {
			log.Fatal(err)
		}
	}

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Gracefully stop the consumers.
	for _, c := range consumers {
		c.Stop()
	}
}
