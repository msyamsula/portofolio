package persistent

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Repository interface {
	GetShortUrl(c context.Context, longUrl string) (string, error)
	SaveShortUrl(c context.Context, shortUrl, longUrl string) error
	GetLongUrl(c context.Context, shortUrl string) (string, error)
}

type Config struct {
	Username string
	Name     string
	Password string
	Host     string
	Port     string
}

func New(config Config) Repository {
	// postgre
	sslmode := "require"
	if os.Getenv("ENVIRONMENT") != "production" {
		// disable for dev
		sslmode = "disable"
	}
	connectionString := fmt.Sprintf("user=%s dbname=%s sslmode=%s password=%s host=%s port=%s",
		config.Username, config.Name, sslmode, config.Password, config.Host, config.Port,
	)
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		log.Fatalln(err)
	}
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
