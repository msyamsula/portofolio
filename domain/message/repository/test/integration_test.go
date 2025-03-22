//go:build integration

package test

import (
	"context"

	"github.com/msyamsula/portofolio/domain/message/repository"
)

func (s *RepositoryTestSuite) TestIntegrationMessage() {
	var err error
	msg := repository.Message{
		Id:         0,
		SenderId:   15,
		ReceiverId: 16,
		Text:       "integration test",
	}
	msg, err = s.realConnection.AddMessage(context.Background(), msg)
	s.Nil(err)
	s.NotZero(msg)

	var msgs []repository.Message
	msgs, err = s.realConnection.GetConversation(context.Background(), msg.SenderId, msg.ReceiverId)
	s.Nil(err)
	s.NotZero(msgs)

	msgs, err = s.realConnection.GetConversation(context.Background(), 1000, 6000)
	s.Nil(err)
	s.Empty(msgs)
}
