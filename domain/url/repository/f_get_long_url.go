package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

var queryGetLongUrl = "SELECT long_url FROM url WHERE short_url = :short_url"

func (repo *Repository) getPersistence(c context.Context, shortUrl string) (string, error) {
	// persistence
	ctx, span := otel.Tracer("").Start(c, "db.transactions")
	defer span.End()

	tx := repo.persistence.MustBeginTx(ctx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})

	var err error
	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(ctx, queryGetLongUrl)
	if err != nil {
		return "", err
	}

	dest := struct {
		LongUrl string `db:"long_url,omitempty"`
	}{}
	dbCtx, dbSpan := otel.Tracer("").Start(ctx, "db.getExecution")
	defer dbSpan.End()
	err = stmt.GetContext(dbCtx, &dest, map[string]interface{}{
		"short_url": shortUrl,
	})
	if err != nil {
		return "", err
	}

	return dest.LongUrl, nil
}

func (repo *Repository) getCache(c context.Context, shortUrl string) (string, error) {
	ctx, span := otel.Tracer("").Start(c, "redis.getCache")
	defer span.End()

	cmd := repo.cache.Get(ctx, shortUrl)

	return cmd.Result()
}

func (repo *Repository) GetLongUrl(c context.Context, shortUrl string) (string, error) {
	ctx, span := otel.Tracer("").Start(c, "repository.GetLongUrl")
	defer span.End()

	var err error
	var longUrl string
	longUrl, err = repo.getCache(ctx, shortUrl)
	if err == nil && longUrl != "" {
		// cached hit
		return longUrl, nil
	}

	// cache miss, get to persistence db
	longUrl, err = repo.getPersistence(ctx, shortUrl)
	if err != nil {
		return longUrl, err
	}

	// set to redis if db success
	repo.setCache(ctx, shortUrl, longUrl)
	return longUrl, nil

}
