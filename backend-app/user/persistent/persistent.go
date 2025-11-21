package persistent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Repository interface {
	InsertUser(c context.Context, user User) (User, error)
	GetUser(c context.Context, username string) (User, error)
}

func NewPostgres(config PostgresConfig) Repository {
	// postgre
	connectionString := fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s host=%s port=%s",
		config.Username, config.DbName, config.Password, config.Host, config.Port,
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
	return &postgres{
		DB: db,
	}
}
