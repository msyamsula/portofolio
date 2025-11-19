package persistent

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

type postgres struct {
	db *sqlx.DB
}

func (repo *postgres) GetShortUrl(c context.Context, longUrl string) (string, error) {
	// persistence
	ctx, span := otel.Tracer("").Start(c, "db.transactions")
	defer span.End()

	tx := repo.db.MustBeginTx(ctx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})

	var err error
	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(ctx, queryGetShortUrl)
	if err != nil {
		return "", err
	}

	dest := ""
	dbCtx, dbSpan := otel.Tracer("").Start(ctx, "db.getExecution")
	defer dbSpan.End()
	err = stmt.GetContext(dbCtx, &dest, map[string]interface{}{
		"long_url": longUrl,
	})
	if err != nil {
		return "", err
	}

	return dest, nil

}

func (repo *postgres) GetLongUrl(c context.Context, shortUrl string) (string, error) {
	// persistence
	ctx, span := otel.Tracer("").Start(c, "db.transactions")
	defer span.End()

	tx := repo.db.MustBeginTx(ctx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})

	var err error
	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(ctx, queryGetLongUrl)
	if err != nil {
		return "", err
	}

	dest := ""
	dbCtx, dbSpan := otel.Tracer("").Start(ctx, "db.getExecution")
	defer dbSpan.End()
	err = stmt.GetContext(dbCtx, &dest, map[string]interface{}{
		"short_url": shortUrl,
	})
	if err != nil {
		return "", err
	}

	return dest, nil

}

func (repo *postgres) SaveShortUrl(c context.Context, shortUrl, longUrl string) error {
	var err error
	ctx, span := otel.Tracer("").Start(c, "db.transaction")
	defer span.End()

	tx := repo.db.MustBeginTx(ctx, &sql.TxOptions{
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
