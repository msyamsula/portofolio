package persistent

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Repository interface {
	GetShortUrl(c context.Context, longUrl string) (string, error)
	SaveShortUrl(c context.Context, shortUrl, longUrl string) error
	GetLongUrl(c context.Context, shortUrl string) (string, error)
}

func NewPostgres(config PostgresConfig) Repository {
	// postgre
	sslmode := "require"
	if os.Getenv("ENVIRONMENT") != "production" {
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
