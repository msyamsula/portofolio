package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/msyamsula/portofolio/domain/message/repository"
	"github.com/nsqio/go-nsq"
	"go.opentelemetry.io/otel"
)

const (
	ConfigReadMessage  = "cfg_read_message"
	TopicReadMessage   = "read_message"
	ChannelReadMessage = "ch_read_message"
)

type ReadMessageHandler struct {
	Repository *repository.Persistence
}

func (s *ReadMessageHandler) HandleMessage(msg *nsq.Message) error {
	ctx, span := otel.Tracer("").Start(context.Background(), "nsqHandler.readMessage")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			fmt.Println(err)
			msg.Requeue(-1)
		} else {
			msg.Finish()
		}
	}()

	type m struct {
		SenderId   int64 `json:"senderId"`
		ReceiverId int64 `json:"receiverId"`
	}
	data := m{}
	err = json.Unmarshal(msg.Body, &data)
	if err != nil {
		fmt.Println(err)
		err = nil // non requeued
		return nil
	}

	err = s.Repository.ReadMessage(ctx, data.SenderId, data.ReceiverId)
	if err != nil {
		return err
	}

	fmt.Println("message read", err, data)
	return nil
}
