package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/msyamsula/portofolio/backend-app/domain/url-shortener/dto"
	"github.com/msyamsula/portofolio/backend-app/mock"
)

func TestRepository_Save_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDatabase(ctrl)
	mockCache := mock.NewMockCache(ctrl)

	repo := NewRepository(mockDB, mockCache)
	ctx := context.Background()

	// Expect DB insert
	mockDB.EXPECT().ExecContext(ctx, gomock.Any(), "test123", "https://example.com").Return(nil, nil)

	// Expect cache invalidation for shortCode
	mockCache.EXPECT().Del(ctx, "test123").Return(redis.NewIntCmd(ctx, 1))

	// Expect cache invalidation for longURL
	mockCache.EXPECT().Del(ctx, "https://example.com").Return(redis.NewIntCmd(ctx, 1))

	err := repo.Save(ctx, "test123", "https://example.com")
	assert.NoError(t, err)
}

func TestRepository_Save_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDatabase(ctrl)
	mockCache := mock.NewMockCache(ctrl)

	repo := NewRepository(mockDB, mockCache)
	ctx := context.Background()

	expectedErr := errors.New("database error")
	mockDB.EXPECT().ExecContext(ctx, gomock.Any(), "test123", "https://example.com").Return(nil, expectedErr)

	err := repo.Save(ctx, "test123", "https://example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save url mapping")
}

func TestRepository_FindByShortCode_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDatabase(ctrl)
	mockCache := mock.NewMockCache(ctrl)

	repo := NewRepository(mockDB, mockCache)
	ctx := context.Background()

	expectedRecord := &dto.URLRecord{
		ShortCode: "test123",
		LongURL:   "https://example.com",
	}

	// Create a mock StringCmd with cached data
	data := []byte(`{"short_code":"test123","long_url":"https://example.com","created_at":1234567890}`)
	cmd := redis.NewStringCmd(ctx)
	cmd.SetVal(string(data))

	mockCache.EXPECT().Get(ctx, "test123").Return(cmd)

	// DB should not be called on cache hit
	result, err := repo.FindByShortCode(ctx, "test123")
	assert.NoError(t, err)
	assert.Equal(t, expectedRecord.ShortCode, result.ShortCode)
	assert.Equal(t, expectedRecord.LongURL, result.LongURL)
}

func TestRepository_FindByShortCode_CacheMiss_DBHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDatabase(ctrl)
	mockCache := mock.NewMockCache(ctrl)

	repo := NewRepository(mockDB, mockCache)
	ctx := context.Background()

	// Cache miss - return nil data with error
	cacheCmd := redis.NewStringCmd(ctx)
	cacheCmd.SetErr(errors.New("cache miss"))
	mockCache.EXPECT().Get(ctx, "test123").Return(cacheCmd)

	// DB hit - return the record
	expectedRecord := &dto.URLRecord{
		ShortCode: "test123",
		LongURL:   "https://example.com",
		CreatedAt: 1234567890,
	}

	mockDB.EXPECT().GetContext(ctx, gomock.Any(), gomock.Any(), "test123").Do(func(_ context.Context, dest *dto.URLRecord, _ interface{}, _ interface{}) {
		*dest = *expectedRecord
	}).Return(nil)

	// Expect cache to be populated
	mockCache.EXPECT().Set(ctx, "test123", gomock.Any(), gomock.Any()).Return(redis.NewStatusCmd(ctx))

	result, err := repo.FindByShortCode(ctx, "test123")
	assert.NoError(t, err)
	assert.Equal(t, expectedRecord.ShortCode, result.ShortCode)
	assert.Equal(t, expectedRecord.LongURL, result.LongURL)
}

func TestRepository_FindByShortCode_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDatabase(ctrl)
	mockCache := mock.NewMockCache(ctrl)

	repo := NewRepository(mockDB, mockCache)
	ctx := context.Background()

	// Cache miss
	cacheCmd := redis.NewStringCmd(ctx)
	cacheCmd.SetErr(errors.New("cache miss"))
	mockCache.EXPECT().Get(ctx, "nonexistent").Return(cacheCmd)

	// DB miss
	mockDB.EXPECT().GetContext(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(sql.ErrNoRows)

	_, err := repo.FindByShortCode(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "url not found")
}

func TestRepository_FindByLongURL_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDatabase(ctrl)
	mockCache := mock.NewMockCache(ctrl)

	repo := NewRepository(mockDB, mockCache)
	ctx := context.Background()

	expectedRecord := &dto.URLRecord{
		ShortCode: "test123",
		LongURL:   "https://example.com",
	}

	// Create a mock StringCmd with cached data
	data := []byte(`{"short_code":"test123","long_url":"https://example.com","created_at":1234567890}`)
	cmd := redis.NewStringCmd(ctx)
	cmd.SetVal(string(data))

	mockCache.EXPECT().Get(ctx, "https://example.com").Return(cmd)

	result, err := repo.FindByLongURL(ctx, "https://example.com")
	assert.NoError(t, err)
	assert.Equal(t, expectedRecord.ShortCode, result.ShortCode)
	assert.Equal(t, expectedRecord.LongURL, result.LongURL)
}

func TestRepository_FindByLongURL_CacheMiss_DBHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDatabase(ctrl)
	mockCache := mock.NewMockCache(ctrl)

	repo := NewRepository(mockDB, mockCache)
	ctx := context.Background()

	// Cache miss
	cacheCmd := redis.NewStringCmd(ctx)
	cacheCmd.SetErr(errors.New("cache miss"))
	mockCache.EXPECT().Get(ctx, "https://example.com").Return(cacheCmd)

	// DB hit
	expectedRecord := &dto.URLRecord{
		ShortCode: "test123",
		LongURL:   "https://example.com",
		CreatedAt: 1234567890,
	}

	mockDB.EXPECT().GetContext(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Do(func(_ context.Context, dest *dto.URLRecord, _ interface{}, _ interface{}) {
		*dest = *expectedRecord
	}).Return(nil)

	// Expect cache to be populated
	mockCache.EXPECT().Set(ctx, "https://example.com", gomock.Any(), gomock.Any()).Return(redis.NewStatusCmd(ctx))

	result, err := repo.FindByLongURL(ctx, "https://example.com")
	assert.NoError(t, err)
	assert.Equal(t, expectedRecord.ShortCode, result.ShortCode)
	assert.Equal(t, expectedRecord.LongURL, result.LongURL)
}

func TestRepository_FindByLongURL_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDatabase(ctrl)
	mockCache := mock.NewMockCache(ctrl)

	repo := NewRepository(mockDB, mockCache)
	ctx := context.Background()

	// Cache miss
	cacheCmd := redis.NewStringCmd(ctx)
	cacheCmd.SetErr(errors.New("cache miss"))
	mockCache.EXPECT().Get(ctx, "https://example.com/notfound").Return(cacheCmd)

	// DB miss
	mockDB.EXPECT().GetContext(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(sql.ErrNoRows)

	_, err := repo.FindByLongURL(ctx, "https://example.com/notfound")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "url not found")
}

func TestNewRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDatabase(ctrl)
	mockCache := mock.NewMockCache(ctrl)

	repo := NewRepository(mockDB, mockCache)

	assert.NotNil(t, repo)
	// repo implements Repository by design
}
