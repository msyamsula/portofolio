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
}
