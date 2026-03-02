package handler

import (
"bytes"
"encoding/json"
"errors"
"net/http"
"net/http/httptest"
"testing"

"github.com/golang/mock/gomock"
"github.com/gorilla/mux"
"github.com/msyamsula/portofolio/backend-app/domain/friend/dto"
"github.com/msyamsula/portofolio/backend-app/mock"
"github.com/stretchr/testify/suite"
)

type FriendHandlerTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	mockSvc *mock.MockFriendService
	handler *Handler
	router  *mux.Router
}

func (s *FriendHandlerTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockSvc = mock.NewMockFriendService(s.ctrl)
	s.handler = New(s.mockSvc)
	s.router = mux.NewRouter()
	s.handler.RegisterRoutes(s.router)
}

func (s *FriendHandlerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *FriendHandlerTestSuite) TestAddFriend_Success() {
	reqBody := dto.AddFriendRequest{SmallID: 1, BigID: 2}
	body, _ := json.Marshal(reqBody)
	s.mockSvc.EXPECT().AddFriend(gomock.Any(), dto.User{ID: 1}, dto.User{ID: 2}).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

func (s *FriendHandlerTestSuite) TestAddFriend_InvalidBody() {
	req := httptest.NewRequest(http.MethodPost, "/add", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *FriendHandlerTestSuite) TestAddFriend_ServiceError() {
	reqBody := dto.AddFriendRequest{SmallID: 1, BigID: 2}
	body, _ := json.Marshal(reqBody)
	s.mockSvc.EXPECT().AddFriend(gomock.Any(), dto.User{ID: 1}, dto.User{ID: 2}).Return(errors.New("service error"))

	req := httptest.NewRequest(http.MethodPost, "/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusInternalServerError, rec.Code)
}

func (s *FriendHandlerTestSuite) TestGetFriends_Success() {
	expected := []dto.User{{ID: 2, Username: "bob"}}
	s.mockSvc.EXPECT().GetFriends(gomock.Any(), dto.User{ID: 1}).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/get?id=1", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

func (s *FriendHandlerTestSuite) TestGetFriends_MissingID() {
	req := httptest.NewRequest(http.MethodGet, "/get", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *FriendHandlerTestSuite) TestGetFriends_InvalidID() {
	req := httptest.NewRequest(http.MethodGet, "/get?id=abc", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *FriendHandlerTestSuite) TestGetFriends_ServiceError() {
	s.mockSvc.EXPECT().GetFriends(gomock.Any(), dto.User{ID: 1}).Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/get?id=1", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusInternalServerError, rec.Code)
}

func (s *FriendHandlerTestSuite) TestNew_ReturnsHandler() {
	h := New(s.mockSvc)
	s.NotNil(h)
}

func (s *FriendHandlerTestSuite) TestRegisterRoutes() {
	router := mux.NewRouter()
	s.handler.RegisterRoutes(router)
	s.NotNil(router)
}

func TestFriendHandlerSuite(t *testing.T) {
	suite.Run(t, new(FriendHandlerTestSuite))
}
