package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"

	"github.com/msyamsula/portofolio/backend-app/domain/url-shortener/dto"
	"github.com/msyamsula/portofolio/backend-app/mock"
)

// URLShortenerRepositoryTestSuite defines the test suite for URL shortener repository
type URLShortenerRepositoryTestSuite struct {
	suite.Suite
	ctrl      *gomock.Controller
	mockDB    *mock.MockDatabase
	mockCache *mock.MockCache
	repo      Repository
	ctx       context.Context
}

func (s *URLShortenerRepositoryTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockDB = mock.NewMockDatabase(s.ctrl)
	s.mockCache = mock.NewMockCache(s.ctrl)
	s.repo = NewRepository(s.mockDB, s.mockCache)
	s.ctx = context.Background()
}

func (s *URLShortenerRepositoryTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// --- Save tests ---

func (s *URLShortenerRepositoryTestSuite) TestSave_Success() {
	s.mockDB.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "test123", "https://example.com").Return(nil, nil)
	s.mockCache.EXPECT().Del(gomock.Any(), "test123").Return(redis.NewIntCmd(s.ctx, 1))
	s.mockCache.EXPECT().Del(gomock.Any(), "https://example.com").Return(redis.NewIntCmd(s.ctx, 1))

	err := s.repo.Save(s.ctx, "test123", "https://example.com")
	s.NoError(err)
}

func (s *URLShortenerRepositoryTestSuite) TestSave_DBError() {
	expectedErr := errors.New("database error")
	s.mockDB.EXPECT().ExecContext(gomock.Any(), gomock.Any(), "test123", "https://example.com").Return(nil, expectedErr)

	err := s.repo.Save(s.ctx, "test123", "https://example.com")
	s.Error(err)
	s.Contains(err.Error(), "failed to save url mapping")
}

// --- FindByShortCode tests ---

func (s *URLShortenerRepositoryTestSuite) TestFindByShortCode_CacheHit() {
	createdAt := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	data := []byte(`{"short_code":"test123","long_url":"https://example.com","created_at":"` + createdAt.Format(time.RFC3339Nano) + `"}`)
	cmd := redis.NewStringCmd(s.ctx)
	cmd.SetVal(string(data))

	s.mockCache.EXPECT().Get(gomock.Any(), "test123").Return(cmd)

	result, err := s.repo.FindByShortCode(s.ctx, "test123")
	s.NoError(err)
	s.Equal("test123", result.ShortCode)
	s.Equal("https://example.com", result.LongURL)
}

func (s *URLShortenerRepositoryTestSuite) TestFindByShortCode_CacheMiss_DBHit() {
	cacheCmd := redis.NewStringCmd(s.ctx)
	cacheCmd.SetErr(errors.New("cache miss"))
	s.mockCache.EXPECT().Get(gomock.Any(), "test123").Return(cacheCmd)

	createdAt := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	expectedRecord := &dto.URLRecord{
		ShortCode: "test123",
		LongURL:   "https://example.com",
		CreatedAt: createdAt,
	}

	s.mockDB.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), "test123").Do(func(_ context.Context, dest *dto.URLRecord, _ interface{}, _ interface{}) {
		*dest = *expectedRecord
	}).Return(nil)

	s.mockCache.EXPECT().Set(gomock.Any(), "test123", gomock.Any(), gomock.Any()).Return(redis.NewStatusCmd(s.ctx))

	result, err := s.repo.FindByShortCode(s.ctx, "test123")
	s.NoError(err)
	s.Equal(expectedRecord.ShortCode, result.ShortCode)
	s.Equal(expectedRecord.LongURL, result.LongURL)
}

func (s *URLShortenerRepositoryTestSuite) TestFindByShortCode_NotFound() {
	cacheCmd := redis.NewStringCmd(s.ctx)
	cacheCmd.SetErr(errors.New("cache miss"))
	s.mockCache.EXPECT().Get(gomock.Any(), "nonexistent").Return(cacheCmd)

	s.mockDB.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sql.ErrNoRows)

	_, err := s.repo.FindByShortCode(s.ctx, "nonexistent")
	s.Error(err)
	s.Contains(err.Error(), "url not found")
}

// --- FindByLongURL tests ---

func (s *URLShortenerRepositoryTestSuite) TestFindByLongURL_CacheHit() {
	createdAt := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	data := []byte(`{"short_code":"test123","long_url":"https://example.com","created_at":"` + createdAt.Format(time.RFC3339Nano) + `"}`)
	cmd := redis.NewStringCmd(s.ctx)
	cmd.SetVal(string(data))

	s.mockCache.EXPECT().Get(gomock.Any(), "https://example.com").Return(cmd)

	result, err := s.repo.FindByLongURL(s.ctx, "https://example.com")
	s.NoError(err)
	s.Equal("test123", result.ShortCode)
	s.Equal("https://example.com", result.LongURL)
}

func (s *URLShortenerRepositoryTestSuite) TestFindByLongURL_CacheMiss_DBHit() {
	cacheCmd := redis.NewStringCmd(s.ctx)
	cacheCmd.SetErr(errors.New("cache miss"))
	s.mockCache.EXPECT().Get(gomock.Any(), "https://example.com").Return(cacheCmd)

	createdAt := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	expectedRecord := &dto.URLRecord{
		ShortCode: "test123",
		LongURL:   "https://example.com",
		CreatedAt: createdAt,
	}

	s.mockDB.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(_ context.Context, dest *dto.URLRecord, _ interface{}, _ interface{}) {
		*dest = *expectedRecord
	}).Return(nil)

	s.mockCache.EXPECT().Set(gomock.Any(), "https://example.com", gomock.Any(), gomock.Any()).Return(redis.NewStatusCmd(s.ctx))

	result, err := s.repo.FindByLongURL(s.ctx, "https://example.com")
	s.NoError(err)
	s.Equal(expectedRecord.ShortCode, result.ShortCode)
	s.Equal(expectedRecord.LongURL, result.LongURL)
}

func (s *URLShortenerRepositoryTestSuite) TestFindByLongURL_NotFound() {
	cacheCmd := redis.NewStringCmd(s.ctx)
	cacheCmd.SetErr(errors.New("cache miss"))
	s.mockCache.EXPECT().Get(gomock.Any(), "https://example.com/notfound").Return(cacheCmd)

	s.mockDB.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sql.ErrNoRows)

	_, err := s.repo.FindByLongURL(s.ctx, "https://example.com/notfound")
	s.Error(err)
	s.Contains(err.Error(), "url not found")
}

// --- Constructor test ---

func (s *URLShortenerRepositoryTestSuite) TestNewRepository() {
	repo := NewRepository(s.mockDB, s.mockCache)
	s.NotNil(repo)
}

// Run the test suite
func TestURLShortenerRepositorySuite(t *testing.T) {
	suite.Run(t, new(URLShortenerRepositoryTestSuite))
}
