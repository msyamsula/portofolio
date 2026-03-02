package repository

import (
"context"
"errors"
"testing"

"github.com/golang/mock/gomock"
"github.com/msyamsula/portofolio/backend-app/domain/message/dto"
"github.com/msyamsula/portofolio/backend-app/mock"
"github.com/stretchr/testify/suite"
)

type MessageRepositoryTestSuite struct {
	suite.Suite
	ctrl   *gomock.Controller
	mockDB *mock.MockDatabase
	repo   Repository
	ctx    context.Context
}

func (s *MessageRepositoryTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockDB = mock.NewMockDatabase(s.ctrl)
	s.repo = NewPostgresRepository(s.mockDB)
	s.ctx = context.Background()
}

func (s *MessageRepositoryTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *MessageRepositoryTestSuite) TestInsertMessage_Success() {
	msg := dto.Message{ID: "msg-1", SenderID: 1, ReceiverID: 2, ConversationID: "conv-1", Data: "hello"}
	expected := dto.Message{ID: "msg-1", SenderID: 1, ReceiverID: 2, ConversationID: "conv-1", Data: "hello"}

	s.mockDB.EXPECT().GetContext(s.ctx, gomock.Any(), gomock.Any(), msg.ID, msg.SenderID, msg.ReceiverID, msg.ConversationID, msg.Data).DoAndReturn(
func(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
result := dest.(*dto.Message)
*result = expected
return nil
},
)

	result, err := s.repo.InsertMessage(s.ctx, msg, TableMessages)
	s.NoError(err)
	s.Equal(expected, result)
}

func (s *MessageRepositoryTestSuite) TestInsertMessage_DBError() {
	msg := dto.Message{ID: "msg-1", SenderID: 1, ReceiverID: 2, ConversationID: "conv-1", Data: "hello"}
	dbErr := errors.New("insert failed")

	s.mockDB.EXPECT().GetContext(s.ctx, gomock.Any(), gomock.Any(), msg.ID, msg.SenderID, msg.ReceiverID, msg.ConversationID, msg.Data).Return(dbErr)

	result, err := s.repo.InsertMessage(s.ctx, msg, TableMessages)
	s.Error(err)
	s.Equal(dbErr, err)
	s.Equal(dto.Message{}, result)
}

func (s *MessageRepositoryTestSuite) TestGetConversation_Success() {
	expected := []dto.Message{
		{ID: "msg-1", SenderID: 1, ReceiverID: 2, Data: "hello"},
	}

	s.mockDB.EXPECT().SelectContext(s.ctx, gomock.Any(), gomock.Any(), "conv-1").DoAndReturn(
func(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
messages := dest.(*[]dto.Message)
*messages = expected
return nil
},
)

	result, err := s.repo.GetConversation(s.ctx, "conv-1", TableMessages)
	s.NoError(err)
	s.Equal(expected, result)
}

func (s *MessageRepositoryTestSuite) TestGetConversation_DBError() {
	dbErr := errors.New("select failed")
	s.mockDB.EXPECT().SelectContext(s.ctx, gomock.Any(), gomock.Any(), "conv-1").Return(dbErr)

	result, err := s.repo.GetConversation(s.ctx, "conv-1", TableMessages)
	s.Error(err)
	s.Nil(result)
}

func (s *MessageRepositoryTestSuite) TestGetConversation_EmptyResult() {
	s.mockDB.EXPECT().SelectContext(s.ctx, gomock.Any(), gomock.Any(), "conv-empty").DoAndReturn(
func(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
messages := dest.(*[]dto.Message)
*messages = []dto.Message{}
return nil
},
)

	result, err := s.repo.GetConversation(s.ctx, "conv-empty", TableMessages)
	s.NoError(err)
	s.Empty(result)
}

func (s *MessageRepositoryTestSuite) TestNewPostgresRepository() {
	repo := NewPostgresRepository(s.mockDB)
	s.NotNil(repo)
}

func TestMessageRepositorySuite(t *testing.T) {
	suite.Run(t, new(MessageRepositoryTestSuite))
}
