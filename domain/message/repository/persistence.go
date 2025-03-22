package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/msyamsula/portofolio/tech-stack/postgres"
	"go.opentelemetry.io/otel"
)

type Persistence struct {
	*postgres.Postgres
}

func New(cfg postgres.Config) *Persistence {
	return &Persistence{
		Postgres: postgres.New(cfg),
	}
}

type Message struct {
	Id         int64  `json:"id,omitempty"`
	SenderId   int64  `json:"sender_id,omitempty"`
	ReceiverId int64  `json:"receiver_id,omitempty"`
	Text       string `json:"text,omitempty"`
	CreateTime time.Time
}

func (s *Persistence) AddMessage(c context.Context, msg Message) (Message, error) {
	ctx, span := otel.Tracer("").Start(c, "repository.persistence.AddMessage")
	defer span.End()

	tx := s.MustBeginTx(ctx, nil)
	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
			tx.Rollback()
		}
	}()

	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(ctx, QueryInsertMessage)
	if err != nil {
		return Message{}, err
	}

	var rows *sql.Rows
	rows, err = stmt.QueryContext(ctx, map[string]interface{}{
		"sender_id":   msg.SenderId,
		"receiver_id": msg.ReceiverId,
		"text":        msg.Text,
	})
	if err != nil {
		return Message{}, err
	}

	for rows.Next() {
		err = rows.Scan(&msg.Id)
		if err != nil {
			return Message{}, err
		}
		break
	}

	err = tx.Commit()
	if err != nil {
		return Message{}, err
	}
	return msg, nil
}

func (s *Persistence) GetConversation(c context.Context, senderId, receiverId int64) ([]Message, error) {
	ctx, span := otel.Tracer("").Start(c, "repository.persistence.GetConversation")
	defer span.End()

	var err error
	tx := s.MustBeginTx(ctx, nil)
	defer func() {
		if err != nil {
			span.RecordError(err)
			tx.Rollback()
		}
	}()

	var stmt *sqlx.NamedStmt
	stmt, err = tx.PrepareNamedContext(ctx, QueryGetConversation)
	if err != nil {
		return []Message{}, err
	}

	var rows *sql.Rows
	rows, err = stmt.QueryContext(ctx, map[string]interface{}{
		"sender_id":   senderId,
		"receiver_id": receiverId,
	})
	if err != nil {
		return []Message{}, err
	}

	messages := []Message{}
	for rows.Next() {
		m := Message{}
		scanErr := rows.Scan(&m.Id, &m.SenderId, &m.ReceiverId, &m.Text, &m.CreateTime)
		if scanErr != nil {
			fmt.Println(scanErr.Error())
			continue
		}
		messages = append(messages, m)
	}

	err = tx.Commit()
	if err != nil {
		return []Message{}, err
	}

	return messages, nil
}
