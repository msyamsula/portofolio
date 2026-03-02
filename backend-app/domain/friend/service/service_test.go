package service

import (
"context"
"errors"
"testing"

"github.com/golang/mock/gomock"
"github.com/msyamsula/portofolio/backend-app/domain/friend/dto"
"github.com/msyamsula/portofolio/backend-app/domain/friend/repository"
"github.com/msyamsula/portofolio/backend-app/mock"
"github.com/stretchr/testify/suite"
)

type FriendServiceTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockRepo *mock.MockFriendRepository
	svc      Service
	ctx      context.Context
}

func (s *FriendServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = mock.NewMockFriendRepository(s.ctrl)
	s.svc = New(s.mockRepo)
	s.ctx = context.Background()
}

func (s *FriendServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *FriendServiceTestSuite) TestAddFriend_Success() {
	userA := dto.User{ID: 1, Username: "alice"}
	userB := dto.User{ID: 2, Username: "bob"}
	s.mockRepo.EXPECT().AddFriend(s.ctx, userA, userB).Return(nil)
	err := s.svc.AddFriend(s.ctx, userA, userB)
	s.NoError(err)
}

func (s *FriendServiceTestSuite) TestAddFriend_SameID_ReturnsError() {
	userA := dto.User{ID: 1}
	userB := dto.User{ID: 1}
	err := s.svc.AddFriend(s.ctx, userA, userB)
	s.Error(err)
	s.Equal(repository.ErrIDMustBeDifferent, err)
}

func (s *FriendServiceTestSuite) TestAddFriend_RepositoryError() {
	userA := dto.User{ID: 1}
	userB := dto.User{ID: 2}
	expectedErr := errors.New("database connection error")
	s.mockRepo.EXPECT().AddFriend(s.ctx, userA, userB).Return(expectedErr)
	err := s.svc.AddFriend(s.ctx, userA, userB)
	s.Error(err)
	s.Equal(expectedErr, err)
}

func (s *FriendServiceTestSuite) TestGetFriends_Success() {
	user := dto.User{ID: 1}
	expected := []dto.User{
		{ID: 2, Username: "bob", Online: true},
		{ID: 3, Username: "charlie", Unread: 5},
	}
	s.mockRepo.EXPECT().GetFriends(s.ctx, user).Return(expected, nil)
	result, err := s.svc.GetFriends(s.ctx, user)
	s.NoError(err)
	s.Equal(expected, result)
	s.Len(result, 2)
}

func (s *FriendServiceTestSuite) TestGetFriends_EmptyList() {
	user := dto.User{ID: 1}
	s.mockRepo.EXPECT().GetFriends(s.ctx, user).Return([]dto.User{}, nil)
	result, err := s.svc.GetFriends(s.ctx, user)
	s.NoError(err)
	s.Empty(result)
}

func (s *FriendServiceTestSuite) TestGetFriends_RepositoryError() {
	user := dto.User{ID: 1}
	expectedErr := errors.New("database error")
	s.mockRepo.EXPECT().GetFriends(s.ctx, user).Return(nil, expectedErr)
	result, err := s.svc.GetFriends(s.ctx, user)
	s.Error(err)
	s.Nil(result)
	s.Equal(expectedErr, err)
}

func (s *FriendServiceTestSuite) TestNew_ReturnsServiceInstance() {
	svc := New(s.mockRepo)
	s.NotNil(svc)
}

func TestFriendServiceSuite(t *testing.T) {
	suite.Run(t, new(FriendServiceTestSuite))
}
