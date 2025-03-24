package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/msyamsula/portofolio/domain/message/repository"
	"github.com/msyamsula/portofolio/domain/message/service"
	"github.com/nsqio/go-nsq"
	"go.opentelemetry.io/otel"
)

const (
	ConfigSaveMessage  = "cfg_save_message"
	TopicSaveMessage   = "send_message"
	ChannelSaveMessage = "ch_save_message"
)

type SaveMessageHandler struct {
	Service *service.Service
}

func (s *SaveMessageHandler) HandleMessage(msg *nsq.Message) error {
	ctx, span := otel.Tracer("").Start(context.Background(), "nsqHandler.saveMessage")
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
		SenderId   int64  `json:"senderId"`
		ReceiverId int64  `json:"receiverId"`
		Text       string `json:"text"`
	}
	data := m{}
	err = json.Unmarshal(msg.Body, &data)
	if err != nil {
		fmt.Println(err)
		err = nil // non requeued
		return nil
	}

	msgSent := repository.Message{
		SenderId:   data.SenderId,
		ReceiverId: data.ReceiverId,
		Text:       data.Text,
	}

	msgSent, err = s.Service.AddMessage(ctx, msgSent)
	if err != nil {
		return err
	}

	fmt.Println("success", err, msgSent)
	return nil
}
