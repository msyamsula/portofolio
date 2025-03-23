package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/msyamsula/portofolio/domain/message"
	"github.com/msyamsula/portofolio/domain/message/repository"
	"github.com/msyamsula/portofolio/domain/message/service"
	"github.com/msyamsula/portofolio/domain/telemetry"
	"github.com/msyamsula/portofolio/tech-stack/postgres"
	"github.com/nsqio/go-nsq"
)

func main() {
	appName := "message-consumer"

	// load env
	godotenv.Load(".env")

	telemetry.InitializeTelemetryTracing(appName, os.Getenv("JAEGER_HOST"))
	var err error

	consumers := []*message.Consumer{}
	lookupds := strings.Split(os.Getenv("LOOKUPDS"), ",")

	// start append consumers here
	{
		saveMessageHandler := &message.SaveMessageHandler{
			Service: service.New(service.Dependencies{
				Persistence: repository.New(postgres.Config{
					Username: os.Getenv("POSTGRES_USERNAME"),
					Password: os.Getenv("POSTGRES_PASSWORD"),
					DbName:   os.Getenv("POSTGRES_DB"),
					Host:     os.Getenv("POSTGRES_HOST"),
					Port:     os.Getenv("POSTGRES_PORT"),
				}),
			}),
		}
		consumerSaveMessage, err := message.New(message.Config{
			Name:      message.ConfigSaveMessage,
			Lookupds:  lookupds,
			NsqConfig: nsq.NewConfig(),
		}, saveMessageHandler)
		if err != nil {
			log.Fatal(err)
		}
		consumers = append(consumers, consumerSaveMessage)
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
