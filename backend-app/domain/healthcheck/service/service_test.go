package service

import (
"context"
"testing"
"time"

"github.com/stretchr/testify/suite"
)

type HealthcheckServiceTestSuite struct {
	suite.Suite
	svc Service
	ctx context.Context
}

func (s *HealthcheckServiceTestSuite) SetupTest() {
	s.svc = New()
	s.ctx = context.Background()
}

func (s *HealthcheckServiceTestSuite) TestCheck_ReturnsHealthy() {
	status, err := s.svc.Check(s.ctx)
	s.NoError(err)
	s.Equal("healthy", status)
}

func (s *HealthcheckServiceTestSuite) TestUptime_PositiveValue() {
	time.Sleep(10 * time.Millisecond)
	uptime := s.svc.Uptime()
	s.Greater(uptime, float64(0))
}

func (s *HealthcheckServiceTestSuite) TestNew_ReturnsServiceInstance() {
	svc := New()
	s.NotNil(svc)
}

func TestHealthcheckServiceSuite(t *testing.T) {
	suite.Run(t, new(HealthcheckServiceTestSuite))
}
