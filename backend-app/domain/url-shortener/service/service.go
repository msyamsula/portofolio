package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/msyamsula/portofolio/backend-app/domain/url-shortener/repository"
)

// Service defines the interface for URL shortening operations
type Service interface {
	Shorten(ctx context.Context, longURL string) (string, error)
	Expand(ctx context.Context, shortCode string) (string, error)
}

// service handles URL shortening operations
type service struct {
	baseURL    string
	repository repository.Repository
}

// New creates a new URL shortener service
func New(baseURL string, repo repository.Repository) Service {
	return &service{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		repository: repo,
	}
}

// Shorten accepts a long URL and returns a shortened URL
func (s *service) Shorten(ctx context.Context, longURL string) (string, error) {
	tracer := otel.Tracer("url-shortener-service")
	ctx, span := tracer.Start(ctx, "service.Shorten",
		trace.WithAttributes(
			attribute.String("url.long_url", longURL),
			attribute.Int("url.long_url_length", len(longURL)),
		),
	)
	defer span.End()

	// Check if URL already exists (cache lookup)
	span.AddEvent("checking_existing_mapping")
	existing, err := s.repository.FindByLongURL(ctx, longURL)
	if err == nil {
		span.SetAttributes(
			attribute.String("short_code", existing.ShortCode),
			attribute.Bool("cache_hit", true),
		)
		span.AddEvent("existing_mapping_found")
		return s.baseURL + "/" + existing.ShortCode, nil
	}
	span.SetAttributes(attribute.Bool("cache_hit", false))

	// Generate new short code
	span.AddEvent("generating_short_code")
	shortCode := s.generateShortCode(longURL)
	span.SetAttributes(attribute.String("short_code", shortCode))

	// Save to repository
	span.AddEvent("saving_mapping", trace.WithAttributes(
		attribute.String("db.operation", "insert"),
		attribute.String("db.table", "url_mappings"),
	))
	if err := s.repository.Save(ctx, shortCode, longURL); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save url mapping")
		return "", err
	}

	shortURL := s.baseURL + "/" + shortCode
	span.SetAttributes(attribute.String("url.short_url", shortURL))
	span.SetStatus(codes.Ok, "")
	return shortURL, nil
}

// Expand retrieves the original long URL from a short code
func (s *service) Expand(ctx context.Context, shortCode string) (string, error) {
	tracer := otel.Tracer("url-shortener-service")
	ctx, span := tracer.Start(ctx, "service.Expand",
		trace.WithAttributes(
			attribute.String("short_code", shortCode),
		),
	)
	defer span.End()

	// Query repository for mapping
	span.AddEvent("fetching_mapping", trace.WithAttributes(
		attribute.String("db.operation", "select"),
		attribute.String("db.table", "url_mappings"),
	))
	mapping, err := s.repository.FindByShortCode(ctx, shortCode)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "url mapping not found")
		return "", err
	}

	span.SetAttributes(
		attribute.String("url.long_url", mapping.LongURL),
		attribute.Int("url.long_url_length", len(mapping.LongURL)),
		attribute.Bool("cache_hit", false), // Will be updated by repository
	)
	span.SetStatus(codes.Ok, "")
	return mapping.LongURL, nil
}

// generateShortCode creates a short code from the long URL
func (s *service) generateShortCode(longURL string) string {
	hash := sha256.Sum256([]byte(longURL))
	// Use first 8 bytes of hash, base64 encoded (about 11 characters)
	return base64.URLEncoding.EncodeToString(hash[:8])[:8]
}
