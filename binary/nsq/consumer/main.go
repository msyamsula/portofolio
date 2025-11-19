package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/msyamsula/portofolio/binary/postgres"
	messageconsumer "github.com/msyamsula/portofolio/domain/message/consumer"
	"github.com/msyamsula/portofolio/domain/message/repository"
	"github.com/msyamsula/portofolio/domain/message/service"
	"github.com/nsqio/go-nsq"
)

func main() {

	// load env
	godotenv.Load(".env")

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
	}

	{
		// read message
		readMessageHandler := &messageconsumer.ReadMessageHandler{
			Repository: repository.New(pg),
		}
		consumerReadMessage, err := messageconsumer.New(messageconsumer.Config{
			Name:      messageconsumer.ConfigReadMessage,
			Lookupds:  lookupds,
			NsqConfig: nsq.NewConfig(),
		}, readMessageHandler)
		if err != nil {
			log.Fatal(err)
		}

		consumers = append(consumers, consumerReadMessage)
	}

	{
		// update unread
		updateUnreadHandler := &messageconsumer.UpdateUnreadHandler{
			Repository: repository.New(pg),
		}
		consumerUpdateUnread, err := messageconsumer.New(messageconsumer.Config{
			Name:      messageconsumer.ConfigUpdateUnread,
			Lookupds:  lookupds,
			NsqConfig: nsq.NewConfig(),
		}, updateUnreadHandler)
		if err != nil {
			log.Fatal(err)
		}

		consumers = append(consumers, consumerUpdateUnread)
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
