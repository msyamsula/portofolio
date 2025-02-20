package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

var querySetShortUrl = "INSERT INTO url (short_url, long_url) VALUES (:short_url, :long_url)"

func (repo *Repository) SetShortUrl(ctx context.Context, shortUrl, longUrl string) error {
	tx := repo.persistence.Db.MustBeginTx(ctx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	var err error
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(ctx, querySetShortUrl)
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx, map[string]interface{}{
		"short_url": shortUrl,
		"long_url":  longUrl,
	})
	if err != nil {
		return err
	}

	// redis part, non blocking
	cmd := repo.cache.Set(ctx, shortUrl, longUrl, repo.cache.Ttl)
	cmd.Result()
	return nil
}
