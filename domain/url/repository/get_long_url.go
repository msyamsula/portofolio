package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

var queryGetLongUrl = "SELECT long_url FROM url WHERE short_url = :short_url"

func (repo *Repository) GetLongUrl(c context.Context, shortUrl string) (string, error) {
	ctx, span := otel.Tracer("").Start(c, "repository.GetLongUrl")
	defer span.End()

	var err error
	var longUrl string

	// get from redis
	redisCtx, redisSpan := otel.Tracer("").Start(ctx, "redis.getCache")
	cmd := repo.cache.Get(redisCtx, shortUrl)
	longUrl, err = cmd.Result()
	redisSpan.End()
	if err == nil && longUrl != "" {
		// cached hit
		return longUrl, nil
	}

	// persistence
	tCtx, tSpan := otel.Tracer("").Start(ctx, "db.transactions")
	tx := repo.persistence.Db.MustBeginTx(tCtx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})

	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(tCtx, queryGetLongUrl)
	if err != nil {
		return "", err
	}

	dest := struct {
		LongUrl string `db:"long_url,omitempty"`
	}{}
	dbCtx, dbSpan := otel.Tracer("").Start(tCtx, "db.getExecution")
	err = stmt.GetContext(dbCtx, &dest, map[string]interface{}{
		"short_url": shortUrl,
	})
	dbSpan.End()
	tSpan.End()
	if err != nil {
		return "", err
	}
	longUrl = dest.LongUrl

	// set to redis if db success
	setCtx, setSpan := otel.Tracer("").Start(ctx, "redis.setCache")
	r := repo.cache.Set(setCtx, shortUrl, longUrl, repo.cache.Ttl)
	r.Result()
	setSpan.End()
	return longUrl, nil

}
