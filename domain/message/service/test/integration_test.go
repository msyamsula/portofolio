//go:build integration

package test

import (
	"context"
	"fmt"

	"github.com/msyamsula/portofolio/domain/message/repository"
	"github.com/msyamsula/portofolio/domain/utils"
)

func (s *ServiceTestSuite) TestIntegrationMessage() {

	msg := repository.Message{
		Id:         0,
		SenderId:   19,
		ReceiverId: 17,
		Text:       utils.RandomName(30),
	}
	ctx := context.Background()
	var err error
	msg, err = s.realService.AddMessage(ctx, msg)
	s.Nil(err)
	s.NotZero(msg.Id)
	var msgs []repository.Message
	msgs, err = s.realService.GetConversation(ctx, msg.SenderId, msg.ReceiverId)
	s.Nil(err)
	s.NotEmpty(msgs)
	for _, m := range msgs {
		fmt.Println(m)
	}
}
