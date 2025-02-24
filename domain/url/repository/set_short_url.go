package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

var querySetShortUrl = "INSERT INTO url (short_url, long_url) VALUES (:short_url, :long_url)"

func (repo *Repository) SetShortUrl(c context.Context, shortUrl, longUrl string) error {
	ctx, span := otel.Tracer("").Start(c, "repository.SetShortUrl")
	defer span.End()

	trxCtx, trxSpan := otel.Tracer("").Start(ctx, "db.transaction")
	tx := repo.persistence.Db.MustBeginTx(trxCtx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})

	var err error
	defer func() {
		if err != nil {
			_, rollbackSpan := otel.Tracer("").Start(trxCtx, "db.rollback")
			tx.Rollback()
			rollbackSpan.End()
		} else {
			_, commitSpan := otel.Tracer("").Start(trxCtx, "db.commit")
			tx.Commit()
			commitSpan.End()
		}
		trxSpan.End()

		// if db success, set to redis, non blocking
		if err == nil {
			redisCtx, redisSpan := otel.Tracer("").Start(ctx, "redis.setCache")
			cmd := repo.cache.Set(redisCtx, shortUrl, longUrl, repo.cache.Ttl)
			cmd.Result()
			redisSpan.End()
		}
	}()

	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(trxCtx, querySetShortUrl)
	if err != nil {
		return err
	}

	// trace the execution
	dbCtx, dbSpan := otel.Tracer("").Start(trxCtx, "db.execution")
	_, err = stmt.ExecContext(dbCtx, map[string]interface{}{
		"short_url": shortUrl,
		"long_url":  longUrl,
	})
	dbSpan.End()
	if err != nil {
		return err
	}

	return nil
}
