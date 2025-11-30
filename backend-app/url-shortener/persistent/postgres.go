package persistent

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	_ "github.com/lib/pq"
)

type postgres struct {
	db *sqlx.DB
}

func NewPostgres(config PostgresConfig) Repository {
	// postgre
	sslmode := "require"
	if config.Env != "production" {
		// disable for dev
		sslmode = "disable"
	}

	connectionString := fmt.Sprintf("user=%s dbname=%s sslmode=%s password=%s host=%s port=%s",
		config.Username, config.Name, sslmode, config.Password, config.Host, config.Port,
	)
	tempdb, err := otelsql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalln(err)
	}
	db := sqlx.NewDb(tempdb, "postgres")

	// connection pool
	db.SetMaxIdleConns(3)
	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(5 * time.Second)
	db.SetConnMaxLifetime(-1)
	// return db
	pg := &postgres{
		db: db,
	}

	return pg
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
