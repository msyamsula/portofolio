package repository

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
	AddFriend(c context.Context, userA, userB User) error
	GetFriends(c context.Context, user User) ([]User, error)
}

func NewPostgres(config PostgresConfig) Repository {
	// postgre
	sslmode := "require"
	if os.Getenv("ENVIRONMENT") != "production" {
		// disable for dev
		sslmode = "disable"
	}
	connectionString := fmt.Sprintf("user=%s dbname=%s sslmode=%s password=%s host=%s port=%s",
		config.Username, config.DbName, sslmode, config.Password, config.Host, config.Port,
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
