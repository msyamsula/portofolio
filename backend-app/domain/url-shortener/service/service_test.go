package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/msyamsula/portofolio/backend-app/domain/url-shortener/dto"
	"github.com/msyamsula/portofolio/backend-app/mock"
)

func TestService_Shorten_NewURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockRepository(ctrl)
	service := New("https://short.est", mockRepo)

	ctx := context.Background()
	longURL := "https://example.com/very/long/url"

	// Expect FindByLongURL to return not found
	mockRepo.EXPECT().FindByLongURL(ctx, longURL).Return(nil, errors.New("not found"))

	// Expect Save to be called
	mockRepo.EXPECT().Save(ctx, gomock.Any(), longURL).Return(nil)

	result, err := service.Shorten(ctx, longURL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == "" {
		t.Error("expected short URL, got empty string")
	}
}

func TestService_Shorten_ExistingURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockRepository(ctrl)
	service := New("https://short.est", mockRepo)

	ctx := context.Background()
	longURL := "https://example.com/very/long/url"
	existingRecord := &dto.URLRecord{
		ShortCode: "existing",
		LongURL:   longURL,
	}

	// Expect FindByLongURL to return existing record
	mockRepo.EXPECT().FindByLongURL(ctx, longURL).Return(existingRecord, nil)

	result, err := service.Shorten(ctx, longURL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "https://short.est/existing"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestService_Shorten_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockRepository(ctrl)
	service := New("https://short.est", mockRepo)

	ctx := context.Background()
	longURL := "https://example.com/very/long/url"

	// Expect FindByLongURL to return not found
	mockRepo.EXPECT().FindByLongURL(ctx, longURL).Return(nil, errors.New("not found"))

	// Expect Save to return error
	mockRepo.EXPECT().Save(ctx, gomock.Any(), longURL).Return(errors.New("database error"))

	_, err := service.Shorten(ctx, longURL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "database error" {
		t.Errorf("expected 'database error', got %v", err)
	}
}

func TestService_Expand_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockRepository(ctrl)
	service := New("https://short.est", mockRepo)

	ctx := context.Background()
	shortCode := "abc12345"
	longURL := "https://example.com/very/long/url"
	record := &dto.URLRecord{
		ShortCode: shortCode,
		LongURL:   longURL,
	}

	mockRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(record, nil)

	result, err := service.Expand(ctx, shortCode)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != longURL {
		t.Errorf("expected %s, got %s", longURL, result)
	}
}

func TestService_Expand_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockRepository(ctrl)
	service := New("https://short.est", mockRepo)

	ctx := context.Background()
	shortCode := "nonexistent"

	mockRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(nil, errors.New("not found"))

	_, err := service.Expand(ctx, shortCode)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
