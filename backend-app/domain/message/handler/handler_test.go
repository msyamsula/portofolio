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
"github.com/msyamsula/portofolio/backend-app/domain/message/dto"
"github.com/msyamsula/portofolio/backend-app/mock"
"github.com/stretchr/testify/suite"
)

type MessageHandlerTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	mockSvc *mock.MockMessageService
	handler *Handler
	router  *mux.Router
}

func (s *MessageHandlerTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockSvc = mock.NewMockMessageService(s.ctrl)
	s.handler = New(s.mockSvc)
	s.router = mux.NewRouter()
	s.handler.RegisterRoutes(s.router)
}

func (s *MessageHandlerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *MessageHandlerTestSuite) TestInsertMessage_Success() {
	reqBody := dto.InsertMessageRequest{SenderID: 1, ReceiverID: 2, ConversationID: "conv-1", Data: "hello"}
	body, _ := json.Marshal(reqBody)

	expected := dto.Message{ID: "msg-1", SenderID: 1, ReceiverID: 2, ConversationID: "conv-1", Data: "hello"}
	s.mockSvc.EXPECT().InsertMessage(gomock.Any(), dto.Message{
		SenderID: 1, ReceiverID: 2, ConversationID: "conv-1", Data: "hello",
	}).Return(expected, nil)

	req := httptest.NewRequest(http.MethodPost, "/insert", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

func (s *MessageHandlerTestSuite) TestInsertMessage_InvalidBody() {
	req := httptest.NewRequest(http.MethodPost, "/insert", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *MessageHandlerTestSuite) TestInsertMessage_ServiceError() {
	reqBody := dto.InsertMessageRequest{SenderID: 1, ReceiverID: 2, ConversationID: "conv-1", Data: "hello"}
	body, _ := json.Marshal(reqBody)

	s.mockSvc.EXPECT().InsertMessage(gomock.Any(), gomock.Any()).Return(dto.Message{}, errors.New("service error"))

	req := httptest.NewRequest(http.MethodPost, "/insert", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusInternalServerError, rec.Code)
}

func (s *MessageHandlerTestSuite) TestGetConversation_Success() {
	expected := []dto.Message{{ID: "msg-1", SenderID: 1, ReceiverID: 2, Data: "hello"}}
	s.mockSvc.EXPECT().GetConversation(gomock.Any(), "conv-1").Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/conversation?conversation_id=conv-1", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

func (s *MessageHandlerTestSuite) TestGetConversation_MissingConversationID() {
	req := httptest.NewRequest(http.MethodGet, "/conversation", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *MessageHandlerTestSuite) TestGetConversation_ServiceError() {
	s.mockSvc.EXPECT().GetConversation(gomock.Any(), "conv-1").Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/conversation?conversation_id=conv-1", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)

	s.Equal(http.StatusInternalServerError, rec.Code)
}

func (s *MessageHandlerTestSuite) TestNew_ReturnsHandler() {
	h := New(s.mockSvc)
	s.NotNil(h)
}

func (s *MessageHandlerTestSuite) TestRegisterRoutes() {
	router := mux.NewRouter()
	s.handler.RegisterRoutes(router)
	s.NotNil(router)
}

func TestMessageHandlerSuite(t *testing.T) {
	suite.Run(t, new(MessageHandlerTestSuite))
}
