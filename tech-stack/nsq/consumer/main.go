package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	messageconsumer "github.com/msyamsula/portofolio/domain/message/consumer"
	"github.com/msyamsula/portofolio/domain/message/repository"
	"github.com/msyamsula/portofolio/domain/message/service"
	"github.com/msyamsula/portofolio/tech-stack/postgres"
	"github.com/msyamsula/portofolio/tech-stack/telemetry"
	"github.com/nsqio/go-nsq"
)

func main() {
	appName := "consumer"

	// load env
	godotenv.Load(".env")

	telemetry.InitializeTelemetryTracing(appName, os.Getenv("JAEGER_HOST"))
	var err error

	consumers := []*messageconsumer.Consumer{}
	lookupds := strings.Split(os.Getenv("LOOKUPDS"), ",")

	pg := postgres.New(postgres.Config{
		Username: os.Getenv("POSTGRES_USERNAME"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DbName:   os.Getenv("POSTGRES_DB"),
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
	})
	// start append consumers here
	{
		saveMessageHandler := &messageconsumer.SaveMessageHandler{
			Service: service.New(service.Dependencies{
				Persistence: repository.New(pg),
			}),
		}
		consumerSaveMessage, err := messageconsumer.New(messageconsumer.Config{
			Name:      messageconsumer.ConfigSaveMessage,
			Lookupds:  lookupds,
			NsqConfig: nsq.NewConfig(),
		}, saveMessageHandler)
		if err != nil {
			log.Fatal(err)
		}
		consumers = append(consumers, consumerSaveMessage)

		readMessageHandler := &messageconsumer.ReadMessageHandler{
			Repository: repository.New(pg),
		}
		consumerReadMessage, err := messageconsumer.New(messageconsumer.Config{
			Name:      messageconsumer.ConfigReadMessage,
			Lookupds:  lookupds,
			NsqConfig: nsq.NewConfig(),
		}, readMessageHandler)

		consumers = append(consumers, consumerReadMessage)
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
