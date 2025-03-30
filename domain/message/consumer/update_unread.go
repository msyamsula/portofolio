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
	ConfigUpdateUnread  = "cfg_update_unread"
	TopicUpdateUnread   = "update_unread"
	ChannelUpdateUnread = "ch_update_unread"
)

type UpdateUnreadHandler struct {
	Repository *repository.Persistence
}

func (s *UpdateUnreadHandler) HandleMessage(msg *nsq.Message) error {
	ctx, span := otel.Tracer("").Start(context.Background(), "nsqHandler.updateUnread")
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
		Unread     int64 `json:"unread"`
	}
	data := m{}
	err = json.Unmarshal(msg.Body, &data)
	if err != nil {
		err = nil // non requeued
		return nil
	}

	err = s.Repository.UpdateUnread(ctx, data.SenderId, data.ReceiverId, data.Unread)
	if err != nil {
		return err
	}

	fmt.Println("success update unread", err, data)
	return nil
}
