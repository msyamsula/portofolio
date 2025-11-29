package persistent

import (
	"context"
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type postgres struct {
	db *sqlx.DB
}

func (repo *postgres) GetShortUrl(c context.Context, longUrl string) (string, error) {
	// persistence
	var err error
	ctx, span := otel.Tracer("persistence").Start(c, "postgres GetShortUrl")
	tx := repo.db.MustBeginTx(ctx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tx.Rollback()
		}
		span.End()
	}()

	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(ctx, queryGetShortUrl)
	if err != nil {
		return "", err
	}

	dest := ""
	err = stmt.GetContext(ctx, &dest, map[string]interface{}{
		"long_url": longUrl,
	})
	if err != nil {
		log.Printf("postgres: %s", err.Error())
		return "", err
	}

	err = tx.Commit()
	return dest, err

}

func (repo *postgres) GetLongUrl(c context.Context, shortUrl string) (string, error) {
	// persistence
	var err error
	ctx, span := otel.Tracer("persistence").Start(c, "GetLongUrl")
	tx := repo.db.MustBeginTx(ctx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.KeyValue{
				Key:   "status",
				Value: attribute.StringValue("rollback"),
			})
			tx.Rollback()
		}
		span.End()
	}()

	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(ctx, queryGetLongUrl)
	if err != nil {
		return "", err
	}

	dest := ""
	err = stmt.GetContext(ctx, &dest, map[string]interface{}{
		"short_url": shortUrl,
	})
	if err != nil {
		return "", err
	}

	err = tx.Commit()
	return dest, err

}

func (repo *postgres) SaveShortUrl(c context.Context, shortUrl, longUrl string) error {
	var err error
	ctx, span := otel.Tracer("persistence").Start(c, "postgres SaveShortUrl")
	tx := repo.db.MustBeginTx(ctx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})

	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tx.Rollback()
		}
		span.End()
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

	err = tx.Commit()
	return err
}
