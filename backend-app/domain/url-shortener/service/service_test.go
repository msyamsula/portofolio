package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/msyamsula/portofolio/backend-app/domain/url-shortener/dto"
	"github.com/msyamsula/portofolio/backend-app/mock"
	"github.com/stretchr/testify/suite"
)

// URLShortenerServiceTestSuite defines the test suite for URL shortener service
type URLShortenerServiceTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockRepo *mock.MockURLShortenerRepository
	svc      Service
	ctx      context.Context
}

func (s *URLShortenerServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = mock.NewMockURLShortenerRepository(s.ctrl)
	s.svc = New("https://short.est", s.mockRepo)
	s.ctx = context.Background()
}

func (s *URLShortenerServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// --- Shorten tests ---

func (s *URLShortenerServiceTestSuite) TestShorten_NewURL_Success() {
	longURL := "https://example.com/very/long/url"

	// FindByLongURL returns not found → new URL
	s.mockRepo.EXPECT().FindByLongURL(gomock.Any(), longURL).Return(nil, errors.New("not found"))
	// Save should be called
	s.mockRepo.EXPECT().Save(gomock.Any(), gomock.Any(), longURL).Return(nil)

	result, err := s.svc.Shorten(s.ctx, longURL)
	s.NoError(err)
	s.NotEmpty(result)
	s.Contains(result, "https://short.est/")
}

func (s *URLShortenerServiceTestSuite) TestShorten_ExistingURL_ReturnsCached() {
	longURL := "https://example.com/very/long/url"
	existingRecord := &dto.URLRecord{
		ShortCode: "existing",
		LongURL:   longURL,
	}

	s.mockRepo.EXPECT().FindByLongURL(gomock.Any(), longURL).Return(existingRecord, nil)
	// Save should NOT be called

	result, err := s.svc.Shorten(s.ctx, longURL)
	s.NoError(err)
	s.Equal("https://short.est/existing", result)
}

func (s *URLShortenerServiceTestSuite) TestShorten_SaveError() {
	longURL := "https://example.com/very/long/url"

	s.mockRepo.EXPECT().FindByLongURL(gomock.Any(), longURL).Return(nil, errors.New("not found"))
	s.mockRepo.EXPECT().Save(gomock.Any(), gomock.Any(), longURL).Return(errors.New("database error"))

	_, err := s.svc.Shorten(s.ctx, longURL)
	s.Error(err)
	s.Equal("database error", err.Error())
}

func (s *URLShortenerServiceTestSuite) TestShorten_BaseURLTrailingSlash() {
	// Verify trailing slash is trimmed
	svc := New("https://short.est/", s.mockRepo)
	longURL := "https://example.com"
	existingRecord := &dto.URLRecord{ShortCode: "abc", LongURL: longURL}

	s.mockRepo.EXPECT().FindByLongURL(gomock.Any(), longURL).Return(existingRecord, nil)

	result, err := svc.Shorten(s.ctx, longURL)
	s.NoError(err)
	s.Equal("https://short.est/abc", result)
}

// --- Expand tests ---

func (s *URLShortenerServiceTestSuite) TestExpand_Success() {
	shortCode := "abc12345"
	longURL := "https://example.com/very/long/url"
	record := &dto.URLRecord{ShortCode: shortCode, LongURL: longURL}

	s.mockRepo.EXPECT().FindByShortCode(gomock.Any(), shortCode).Return(record, nil)

	result, err := s.svc.Expand(s.ctx, shortCode)
	s.NoError(err)
	s.Equal(longURL, result)
}

func (s *URLShortenerServiceTestSuite) TestExpand_NotFound() {
	shortCode := "nonexistent"

	s.mockRepo.EXPECT().FindByShortCode(gomock.Any(), shortCode).Return(nil, errors.New("not found"))

	_, err := s.svc.Expand(s.ctx, shortCode)
	s.Error(err)
}

// --- Internal method tests ---

func (s *URLShortenerServiceTestSuite) TestGenerateShortCode_Deterministic() {
	svcImpl := s.svc.(*service)
	code1 := svcImpl.generateShortCode("https://example.com")
	code2 := svcImpl.generateShortCode("https://example.com")
	s.Equal(code1, code2)
}

func (s *URLShortenerServiceTestSuite) TestGenerateShortCode_DifferentInputs() {
	svcImpl := s.svc.(*service)
	code1 := svcImpl.generateShortCode("https://example.com/a")
	code2 := svcImpl.generateShortCode("https://example.com/b")
	s.NotEqual(code1, code2)
}

func (s *URLShortenerServiceTestSuite) TestGenerateShortCode_Length() {
	svcImpl := s.svc.(*service)
	code := svcImpl.generateShortCode("https://example.com")
	s.Len(code, 8)
}

// --- Constructor test ---

func (s *URLShortenerServiceTestSuite) TestNew_ReturnsServiceInstance() {
	svc := New("https://short.est", s.mockRepo)
	s.NotNil(svc)
}

// Run the test suite
func TestURLShortenerServiceSuite(t *testing.T) {
	suite.Run(t, new(URLShortenerServiceTestSuite))
}
