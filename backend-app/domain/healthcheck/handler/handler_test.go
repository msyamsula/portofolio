package handler

import (
"errors"
"net/http"
"net/http/httptest"
"testing"

"github.com/golang/mock/gomock"
"github.com/gorilla/mux"
"github.com/msyamsula/portofolio/backend-app/mock"
"github.com/stretchr/testify/suite"
)

type HealthcheckHandlerTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	mockSvc *mock.MockHealthcheckService
	handler *Handler
	router  *mux.Router
}

func (s *HealthcheckHandlerTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockSvc = mock.NewMockHealthcheckService(s.ctrl)
	s.handler = New(s.mockSvc)
	s.router = mux.NewRouter()
	s.handler.RegisterRoutes(s.router)
}

func (s *HealthcheckHandlerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *HealthcheckHandlerTestSuite) TestCheck_Success() {
	s.mockSvc.EXPECT().Check(gomock.Any()).Return("healthy", nil)
	s.mockSvc.EXPECT().Uptime().Return(42.5)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

func (s *HealthcheckHandlerTestSuite) TestCheck_WithoutTrailingSlash() {
	s.mockSvc.EXPECT().Check(gomock.Any()).Return("healthy", nil)
	s.mockSvc.EXPECT().Uptime().Return(10.0)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

func (s *HealthcheckHandlerTestSuite) TestCheck_ServiceError() {
	s.mockSvc.EXPECT().Check(gomock.Any()).Return("", errors.New("check failed"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusInternalServerError, rec.Code)
}

func (s *HealthcheckHandlerTestSuite) TestNew_ReturnsHandler() {
	h := New(s.mockSvc)
	s.NotNil(h)
}

func (s *HealthcheckHandlerTestSuite) TestRegisterRoutes() {
	router := mux.NewRouter()
	s.handler.RegisterRoutes(router)
	s.NotNil(router)
}

func TestHealthcheckHandlerSuite(t *testing.T) {
	suite.Run(t, new(HealthcheckHandlerTestSuite))
}
