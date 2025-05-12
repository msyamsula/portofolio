package test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/msyamsula/portofolio/binary/postgres"
	"github.com/msyamsula/portofolio/domain/message/repository"
	"github.com/msyamsula/portofolio/domain/message/service"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	suite.SetupAllSuite
	suite.TearDownAllSuite

	mockErr error
	ctrl    *gomock.Controller

	mockPersistence *MockPersistenceLayer

	realService *service.Service
}

func (s *ServiceTestSuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())
	s.mockErr = errors.New("ops")

	s.mockPersistence = NewMockPersistenceLayer(s.ctrl)
	s.realService = &service.Service{
		Persistence: &repository.Persistence{
			Postgres: postgres.New(postgres.Config{
				Username: "admin",
				Password: "admin",
				DbName:   "postgres",
				Host:     "0.0.0.0",
				Port:     "5432",
			}),
		},
	}
}
func (s *ServiceTestSuite) TearDownSuite() {
	s.ctrl.Finish()
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
