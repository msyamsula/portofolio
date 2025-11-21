package persistence

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Persistence interface {
	InsertMessage(c context.Context, msg Message, table string) (Message, error)
	GetReadConversation(c context.Context, conversationId string, table string) ([]Message, error)
	ReadMessage(c context.Context, conversationId string) error
}

func NewPostgres(config PostgresConfig, env string) Persistence {
	// postgre
	sslmode := "disable"
	if env == "production" {
		sslmode = "require"
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
