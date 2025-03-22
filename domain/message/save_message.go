package message

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nsqio/go-nsq"
	"go.opentelemetry.io/otel"
)

const (
	ConfigSaveMessage  = "cfg_save_message"
	TopicSaveMessage   = "send_message"
	ChannelSaveMessage = "ch_save_message"
)

type SaveMessageHandler struct{}

func (s *SaveMessageHandler) HandleMessage(msg *nsq.Message) error {
	_, span := otel.Tracer("").Start(context.Background(), "nsqHandler.saveMessage")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			msg.Requeue(-1)
		} else {
			msg.Finish()
		}
	}()

	type dataType struct {
		Message string `json:"message"`
		Number  int64  `json:"number"`
	}
	data := dataType{}
	err = json.Unmarshal(msg.Body, &data)
	if err != nil {
		err = nil // non requeued
		return nil
	}

	fmt.Println(data)

	return nil
}
