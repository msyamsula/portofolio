package minimal

import (
"testing"

"github.com/stretchr/testify/suite"
)

type ServerTestSuite struct {
suite.Suite
}

func (s *ServerTestSuite) TestNew_CreatesServer() {
srv := New(":8080")
	s.NotNil(srv)
	s.NotNil(srv.router)
	s.NotNil(srv.server)
}

func (s *ServerTestSuite) TestNew_SetsAddress() {
	srv := New(":9090")
	s.Equal(":9090", srv.server.Addr)
}

func (s *ServerTestSuite) TestRouter_ReturnsRouter() {
	srv := New(":8080")
	r := srv.Router()
	s.NotNil(r)
	s.Equal(srv.router, r)
}

func (s *ServerTestSuite) TestNew_SetsTimeouts() {
	srv := New(":8080")
	s.NotZero(srv.server.ReadTimeout)
	s.NotZero(srv.server.WriteTimeout)
	s.NotZero(srv.server.IdleTimeout)
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
