//go:build integration

package test

import (
	"context"
	"fmt"

	"github.com/msyamsula/portofolio/binary/postgres"
	"github.com/msyamsula/portofolio/domain/message/repository"
)

func (s *RepositoryTestSuite) TestIntegrationMessage() {

	realConnection = &repository.Persistence{
		Postgres: postgres.New(postgres.Config{
			Username: "admin",
			Password: "admin",
			DbName:   "postgres",
			Host:     "0.0.0.0",
			Port:     "5432",
		}),
	}

	var err error
	msg := repository.Message{
		Id:         0,
		SenderId:   1000,
		ReceiverId: 2000,
		Text:       "integration test 2",
	}
	msg, err = realConnection.AddMessage(context.Background(), msg)
	s.Nil(err)
	s.NotZero(msg)

	err = realConnection.ReadMessage(context.Background(), msg.SenderId, msg.ReceiverId)
	s.Nil(err)

	var msgs []repository.Message
	msgs, err = realConnection.GetConversation(context.Background(), msg.SenderId, msg.ReceiverId)
	s.Nil(err)
	s.NotZero(msgs)
	for _, m := range msgs {
		fmt.Println(m)
	}

	msgs, err = realConnection.GetConversation(context.Background(), 1000, 6000)
	s.Nil(err)
	s.Empty(msgs)

	err = realConnection.UpdateUnread(context.Background(), 1, 4, 99)
	s.Nil(err)

}
