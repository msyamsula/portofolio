package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

var queryGetLongUrl = "SELECT long_url FROM url WHERE short_url = :short_url"

func (repo *Repository) GetLongUrl(ctx context.Context, shortUrl string) (string, error) {
	var err error
	var longUrl string

	// get from redis
	cmd := repo.cache.Get(ctx, shortUrl)
	longUrl, err = cmd.Result()
	if err == nil && longUrl != "" {
		// cached hit
		return longUrl, nil
	}

	// persistence
	tx := repo.persistence.Db.MustBeginTx(ctx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(ctx, queryGetLongUrl)
	if err != nil {
		return "", err
	}

	dest := struct {
		LongUrl string `db:"long_url,omitempty"`
	}{}
	err = stmt.GetContext(ctx, &dest, map[string]interface{}{
		"short_url": shortUrl,
	})
	if err != nil {
		return "", err
	}

	r := repo.cache.Set(ctx, shortUrl, longUrl, repo.cache.Ttl)
	r.Result()
	return dest.LongUrl, nil

}
