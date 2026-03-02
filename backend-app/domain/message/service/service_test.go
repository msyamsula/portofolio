package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/msyamsula/portofolio/backend-app/domain/message/dto"
	"github.com/msyamsula/portofolio/backend-app/domain/message/repository"
	"github.com/msyamsula/portofolio/backend-app/mock"
	"github.com/stretchr/testify/suite"
)

type MessageServiceTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockRepo *mock.MockMessageRepository
	svc      Service
	ctx      context.Context
}

func (s *MessageServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = mock.NewMockMessageRepository(s.ctrl)
	s.svc = New(s.mockRepo)
	s.ctx = context.Background()
}

func (s *MessageServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *MessageServiceTestSuite) TestInsertMessage_Success() {
	msg := dto.Message{SenderID: 1, ReceiverID: 2, ConversationID: "conv-1", Data: "hello"}
	expected := dto.Message{ID: "msg-1", SenderID: 1, ReceiverID: 2, ConversationID: "conv-1", Data: "hello"}
	s.mockRepo.EXPECT().InsertMessage(s.ctx, msg, repository.TableMessages).Return(expected, nil)

	result, err := s.svc.InsertMessage(s.ctx, msg)
	s.NoError(err)
	s.Equal(expected, result)
}

func (s *MessageServiceTestSuite) TestInsertMessage_SenderIDZero() {
	msg := dto.Message{SenderID: 0, ReceiverID: 2, Data: "hello"}
	result, err := s.svc.InsertMessage(s.ctx, msg)
	s.Error(err)
	s.Equal(repository.ErrBadRequest, err)
	s.Equal(dto.Message{}, result)
}

func (s *MessageServiceTestSuite) TestInsertMessage_NegativeSenderID() {
	msg := dto.Message{SenderID: -1, ReceiverID: 2, Data: "hello"}
	_, err := s.svc.InsertMessage(s.ctx, msg)
	s.Error(err)
	s.Equal(repository.ErrBadRequest, err)
}

func (s *MessageServiceTestSuite) TestInsertMessage_ReceiverIDZero() {
	msg := dto.Message{SenderID: 1, ReceiverID: 0, Data: "hello"}
	result, err := s.svc.InsertMessage(s.ctx, msg)
	s.Error(err)
	s.Equal(repository.ErrBadRequest, err)
	s.Equal(dto.Message{}, result)
}

func (s *MessageServiceTestSuite) TestInsertMessage_SameSenderAndReceiver() {
	msg := dto.Message{SenderID: 1, ReceiverID: 1, Data: "hello"}
	result, err := s.svc.InsertMessage(s.ctx, msg)
	s.Error(err)
	s.Equal(repository.ErrBadRequest, err)
	s.Equal(dto.Message{}, result)
}

func (s *MessageServiceTestSuite) TestInsertMessage_EmptyData() {
	msg := dto.Message{SenderID: 1, ReceiverID: 2, Data: ""}
	result, err := s.svc.InsertMessage(s.ctx, msg)
	s.Error(err)
	s.Equal(repository.ErrBadRequest, err)
	s.Equal(dto.Message{}, result)
}

func (s *MessageServiceTestSuite) TestInsertMessage_RepositoryError() {
	msg := dto.Message{SenderID: 1, ReceiverID: 2, ConversationID: "conv-1", Data: "hello"}
	repoErr := errors.New("database error")
	s.mockRepo.EXPECT().InsertMessage(s.ctx, msg, repository.TableMessages).Return(dto.Message{}, repoErr)

	result, err := s.svc.InsertMessage(s.ctx, msg)
	s.Error(err)
	s.Equal(repoErr, err)
	s.Equal(dto.Message{}, result)
}

func (s *MessageServiceTestSuite) TestGetConversation_Success() {
	expected := []dto.Message{
		{ID: "msg-1", SenderID: 1, ReceiverID: 2, Data: "hello"},
		{ID: "msg-2", SenderID: 2, ReceiverID: 1, Data: "hi"},
	}
	s.mockRepo.EXPECT().GetConversation(s.ctx, "conv-1", repository.TableMessages).Return(expected, nil)

	result, err := s.svc.GetConversation(s.ctx, "conv-1")
	s.NoError(err)
	s.Equal(expected, result)
	s.Len(result, 2)
}

func (s *MessageServiceTestSuite) TestGetConversation_EmptyConversationID() {
	result, err := s.svc.GetConversation(s.ctx, "")
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "conversation_id is required")
}

func (s *MessageServiceTestSuite) TestGetConversation_RepositoryError() {
	repoErr := errors.New("database error")
	s.mockRepo.EXPECT().GetConversation(s.ctx, "conv-1", repository.TableMessages).Return(nil, repoErr)

	result, err := s.svc.GetConversation(s.ctx, "conv-1")
	s.Error(err)
	s.Nil(result)
	s.Equal(repoErr, err)
}

func (s *MessageServiceTestSuite) TestGetConversation_EmptyResult() {
	s.mockRepo.EXPECT().GetConversation(s.ctx, "conv-empty", repository.TableMessages).Return([]dto.Message{}, nil)

	result, err := s.svc.GetConversation(s.ctx, "conv-empty")
	s.NoError(err)
	s.Empty(result)
}

func (s *MessageServiceTestSuite) TestNew_ReturnsServiceInstance() {
	svc := New(s.mockRepo)
	s.NotNil(svc)
}

func TestMessageServiceSuite(t *testing.T) {
	suite.Run(t, new(MessageServiceTestSuite))
}
