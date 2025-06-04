package test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	suite.SetupAllSuite
	suite.TearDownAllSuite

	mockErr error
	ctrl    *gomock.Controller

	mockPersistence *MockPersistenceLayer
}

func (s *ServiceTestSuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())
	s.mockErr = errors.New("ops")

	s.mockPersistence = NewMockPersistenceLayer(s.ctrl)
}
func (s *ServiceTestSuite) TearDownSuite() {
	s.ctrl.Finish()
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
