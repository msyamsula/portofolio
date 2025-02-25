package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

var querySetShortUrl = "INSERT INTO url (short_url, long_url) VALUES (:short_url, :long_url)"

func (repo *Repository) setPersistence(c context.Context, shortUrl, longUrl string) error {
	var err error
	ctx, span := otel.Tracer("").Start(c, "db.transaction")
	defer span.End()

	tx := repo.persistence.MustBeginTx(ctx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})

	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(ctx, querySetShortUrl)
	if err != nil {
		return err
	}
	// trace the execution
	dbCtx, dbSpan := otel.Tracer("").Start(ctx, "db.execution")
	defer dbSpan.End()

	_, err = stmt.ExecContext(dbCtx, map[string]interface{}{
		"short_url": shortUrl,
		"long_url":  longUrl,
	})
	if err != nil {
		return err
	}

	_, commitSpan := otel.Tracer("").Start(ctx, "db.commit")
	defer commitSpan.End()
	err = tx.Commit()
	if err != nil {
		_, rollbackSpan := otel.Tracer("").Start(ctx, "db.rollback")
		defer rollbackSpan.End()
		tx.Rollback()
		return err
	}

	return nil
}

func (repo *Repository) setCache(c context.Context, shortUrl, longUrl string) error {
	redisCtx, redisSpan := otel.Tracer("").Start(c, "redis.setCache")
	defer redisSpan.End()
	cmd := repo.cache.Set(redisCtx, shortUrl, longUrl, repo.cache.Ttl)
	_, err := cmd.Result()
	return err
}

func (repo *Repository) SetShortUrl(c context.Context, shortUrl, longUrl string) error {
	ctx, span := otel.Tracer("").Start(c, "repository.SetShortUrl")
	defer span.End()

	var err error

	err = repo.setPersistence(ctx, shortUrl, longUrl)
	if err != nil {
		return err
	}

	// if db success, set to redis, non blocking
	repo.setCache(ctx, shortUrl, longUrl)
	return nil
}
