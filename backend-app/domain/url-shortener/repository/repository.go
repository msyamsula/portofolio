package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/msyamsula/portofolio/backend-app/domain/url-shortener/dto"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/database/postgres"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/database/redis"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
)

var (
	tableName = "url_shortener"
)

// Repository defines the interface for URL persistence operations
type Repository interface {
	Save(ctx context.Context, shortCode, longURL string) error
	FindByShortCode(ctx context.Context, shortCode string) (*dto.URLRecord, error)
	FindByLongURL(ctx context.Context, longURL string) (*dto.URLRecord, error)
}

// repository implements cache-aside pattern using Redis and PostgreSQL
type repository struct {
	db       postgres.Database
	cache    redis.Cache
	cacheTTL time.Duration
}

// NewRepository creates a new repository with cache-aside pattern
func NewRepository(db postgres.Database, cache redis.Cache) Repository {
	return &repository{
		db:       db,
		cache:    cache,
		cacheTTL: 24 * time.Hour,
	}
}

// Save stores the URL mapping in PostgreSQL and invalidates related cache
func (r *repository) Save(ctx context.Context, shortCode, longURL string) error {
	tracer := otel.Tracer("url-shortener-repository")
	ctx, span := tracer.Start(ctx, "repository.Save",
		trace.WithAttributes(
			attribute.String("short_code", shortCode),
			attribute.String("long_url", longURL),
			attribute.String("db.table", tableName),
			attribute.String("db.operation", "INSERT"),
		),
	)
	defer span.End()

	query := fmt.Sprintf(`
		INSERT INTO %s (short, long)
		VALUES ($1, $2)
	`, tableName)
	span.AddEvent("executing_sql", trace.WithAttributes(
		attribute.String("db.statement", fmt.Sprintf("INSERT INTO %s", tableName)),
	))
	_, err := r.db.ExecContext(ctx, query, shortCode, longURL)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save url mapping")
		return fmt.Errorf("failed to save url mapping: %w", err)
	}
	span.AddEvent("sql_executed_successfully")

	// invalidate after insertion
	span.AddEvent("invalidating_cache")
	if err := r.invalidateCache(ctx, shortCode); err != nil {
		span.AddEvent("cache_invalidation_failed", trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
	}
	if err := r.invalidateCache(ctx, longURL); err != nil {
		span.AddEvent("cache_invalidation_failed", trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// FindByShortCode retrieves URL by short code using cache-aside pattern
func (r *repository) FindByShortCode(ctx context.Context, shortCode string) (*dto.URLRecord, error) {
	tracer := otel.Tracer("url-shortener-repository")
	ctx, span := tracer.Start(ctx, "repository.FindByShortCode",
		trace.WithAttributes(
			attribute.String("short_code", shortCode),
			attribute.String("db.table", tableName),
		),
	)
	defer span.End()

	// 1. Try cache first
	span.AddEvent("cache_lookup", trace.WithAttributes(
		attribute.String("cache.key", shortCode),
	))
	cached, err := r.getFromCache(ctx, shortCode)
	if err == nil {
		span.SetAttributes(attribute.Bool("cache.hit", true))
		span.AddEvent("cache_hit")
		return cached, nil
	}
	span.SetAttributes(attribute.Bool("cache.hit", false))
	span.AddEvent("cache_miss")

	// 2. Cache miss, query database
	span.AddEvent("database_lookup", trace.WithAttributes(
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.statement", fmt.Sprintf("SELECT FROM %s WHERE short = ?", tableName)),
	))
	var mapping dto.URLRecord
	query := fmt.Sprintf(`SELECT short, long, created_at FROM %s WHERE short = $1`, tableName)
	err = r.db.GetContext(ctx, &mapping, query, shortCode)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "database query failed")
		return nil, fmt.Errorf("url not found: %w", err)
	}
	span.AddEvent("database_query_success")

	// 3. Populate cache for next request
	span.AddEvent("cache_write", trace.WithAttributes(
		attribute.String("cache.key", shortCode),
		attribute.String("cache.ttl", r.cacheTTL.String()),
	))
	r.cacheURL(ctx, shortCode, mapping)

	span.SetStatus(codes.Ok, "")
	return &mapping, nil
}

// FindByLongURL retrieves existing mapping by long URL
func (r *repository) FindByLongURL(ctx context.Context, longURL string) (*dto.URLRecord, error) {
	tracer := otel.Tracer("url-shortener-repository")
	ctx, span := tracer.Start(ctx, "repository.FindByLongURL",
		trace.WithAttributes(
			attribute.String("long_url", longURL),
			attribute.String("db.table", tableName),
		),
	)
	defer span.End()

	// 1. Try cache first
	span.AddEvent("cache_lookup", trace.WithAttributes(
		attribute.String("cache.key", longURL),
	))
	cached, err := r.getFromCache(ctx, longURL)
	if err == nil {
		span.SetAttributes(attribute.Bool("cache.hit", true))
		span.AddEvent("cache_hit")
		return cached, nil
	}
	span.SetAttributes(attribute.Bool("cache.hit", false))
	span.AddEvent("cache_miss")

	// Query database
	span.AddEvent("database_lookup", trace.WithAttributes(
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.statement", fmt.Sprintf("SELECT FROM %s WHERE long = ?", tableName)),
	))
	var mapping dto.URLRecord
	query := fmt.Sprintf(`SELECT short, long, created_at FROM %s WHERE long = $1`, tableName)
	err = r.db.GetContext(ctx, &mapping, query, longURL)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "database query failed")
		return nil, fmt.Errorf("url not found: %w", err)
	}
	span.AddEvent("database_query_success")

	// Populate cache
	span.AddEvent("cache_write")
	r.cacheURL(ctx, longURL, mapping)

	span.SetStatus(codes.Ok, "")
	return &mapping, nil
}

// cacheURL stores the URL mapping in Redis
func (r *repository) cacheURL(ctx context.Context, key string, value dto.URLRecord) {
	tracer := otel.Tracer("url-shortener-repository")
	_, span := tracer.Start(ctx, "repository.cacheURL",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.ttl", r.cacheTTL.String()),
			attribute.String("cache.operation", "SET"),
		),
	)
	defer span.End()

	data, _ := json.Marshal(value)
	if err := r.cache.Set(ctx, key, data, r.cacheTTL).Err(); err != nil {
		span.RecordError(err)
		logger.Error("failed to cache url", map[string]any{"shortCode": value.ShortCode, "longURL": value.LongURL, "error": err})
		return
	}
	span.AddEvent("cache_write_success")
}

// getFromCache retrieves URL mapping from Redis
func (r *repository) getFromCache(ctx context.Context, key string) (*dto.URLRecord, error) {
	tracer := otel.Tracer("url-shortener-repository")
	_, span := tracer.Start(ctx, "repository.getFromCache",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "GET"),
		),
	)
	defer span.End()

	data, err := r.cache.Get(ctx, key).Bytes()
	if err != nil {
		span.SetAttributes(attribute.Bool("cache.hit", false))
		return nil, err
	}

	mapping := dto.URLRecord{}
	if err = json.Unmarshal(data, &mapping); err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Bool("cache.hit", true))
	span.AddEvent("cache_hit")
	return &mapping, nil
}

// invalidateCache removes the URL mapping from Redis cache
func (r *repository) invalidateCache(ctx context.Context, key string) error {
	tracer := otel.Tracer("url-shortener-repository")
	_, span := tracer.Start(ctx, "repository.invalidateCache",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.operation", "DEL"),
		),
	)
	defer span.End()

	if err := r.cache.Del(ctx, key).Err(); err != nil {
		span.RecordError(err)
		return err
	}
	span.AddEvent("cache_invalidation_success")
	return nil
}
