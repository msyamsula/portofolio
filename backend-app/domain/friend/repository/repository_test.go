package repository

import (
"context"
"errors"
"testing"

"github.com/golang/mock/gomock"
"github.com/msyamsula/portofolio/backend-app/domain/friend/dto"
"github.com/msyamsula/portofolio/backend-app/mock"
"github.com/stretchr/testify/suite"
)

type FriendRepositoryTestSuite struct {
	suite.Suite
	ctrl   *gomock.Controller
	mockDB *mock.MockDatabase
	repo   Repository
	ctx    context.Context
}

func (s *FriendRepositoryTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockDB = mock.NewMockDatabase(s.ctrl)
	s.repo = NewPostgresRepository(s.mockDB)
	s.ctx = context.Background()
}

func (s *FriendRepositoryTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *FriendRepositoryTestSuite) TestAddFriend_Success() {
	userA := dto.User{ID: 1}
	userB := dto.User{ID: 2}
	s.mockDB.EXPECT().ExecContext(s.ctx, gomock.Any(), int64(1), int64(2)).Return(nil, nil)
	err := s.repo.AddFriend(s.ctx, userA, userB)
	s.NoError(err)
}

func (s *FriendRepositoryTestSuite) TestAddFriend_ReversedOrder() {
	userA := dto.User{ID: 5}
	userB := dto.User{ID: 3}
	// small_id=3, big_id=5
	s.mockDB.EXPECT().ExecContext(s.ctx, gomock.Any(), int64(3), int64(5)).Return(nil, nil)
	err := s.repo.AddFriend(s.ctx, userA, userB)
	s.NoError(err)
}

func (s *FriendRepositoryTestSuite) TestAddFriend_SameID_ReturnsError() {
	userA := dto.User{ID: 1}
	userB := dto.User{ID: 1}
	err := s.repo.AddFriend(s.ctx, userA, userB)
	s.Error(err)
	s.Equal(ErrIDMustBeDifferent, err)
}

func (s *FriendRepositoryTestSuite) TestAddFriend_DBError() {
	userA := dto.User{ID: 1}
	userB := dto.User{ID: 2}
	dbErr := errors.New("connection refused")
	s.mockDB.EXPECT().ExecContext(s.ctx, gomock.Any(), int64(1), int64(2)).Return(nil, dbErr)
	err := s.repo.AddFriend(s.ctx, userA, userB)
	s.Error(err)
	s.Equal(dbErr, err)
}

func (s *FriendRepositoryTestSuite) TestGetFriends_Success() {
	user := dto.User{ID: 1}
	expected := []dto.User{
		{ID: 2, Username: "bob", Online: true},
	}
	s.mockDB.EXPECT().SelectContext(s.ctx, gomock.Any(), gomock.Any(), user.ID).DoAndReturn(
func(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
users := dest.(*[]dto.User)
*users = expected
return nil
},
)
	result, err := s.repo.GetFriends(s.ctx, user)
	s.NoError(err)
	s.Equal(expected, result)
}

func (s *FriendRepositoryTestSuite) TestGetFriends_DBError() {
	user := dto.User{ID: 1}
	dbErr := errors.New("query failed")
	s.mockDB.EXPECT().SelectContext(s.ctx, gomock.Any(), gomock.Any(), user.ID).Return(dbErr)
	result, err := s.repo.GetFriends(s.ctx, user)
	s.Error(err)
	s.Nil(result)
}

func (s *FriendRepositoryTestSuite) TestGetFriends_Empty() {
	user := dto.User{ID: 1}
	s.mockDB.EXPECT().SelectContext(s.ctx, gomock.Any(), gomock.Any(), user.ID).DoAndReturn(
func(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
users := dest.(*[]dto.User)
*users = []dto.User{}
return nil
},
)
	result, err := s.repo.GetFriends(s.ctx, user)
	s.NoError(err)
	s.Empty(result)
}

func (s *FriendRepositoryTestSuite) TestNewPostgresRepository() {
	repo := NewPostgresRepository(s.mockDB)
	s.NotNil(repo)
}

func TestFriendRepositorySuite(t *testing.T) {
	suite.Run(t, new(FriendRepositoryTestSuite))
}
