package test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
	suite.Suite
	suite.SetupAllSuite
	suite.TearDownAllSuite

	mockSvc *MockService
	mockErr error
	ctrl    *gomock.Controller
}

func (s *UserTestSuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())

	s.mockSvc = NewMockService(s.ctrl)
	s.mockErr = errors.New("ops")

}
func (s *UserTestSuite) TearDownSuite() {
	s.ctrl.Finish()
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
