package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

type postgres struct {
	*sqlx.DB
}

func (s *postgres) insertReadMessageByTx(c context.Context, tx *sqlx.Tx, msg Message, table string) (Message, error) {
	ctx, span := otel.Tracer("").Start(c, "repository.persistence.insertReadMessgeByTx")
	defer span.End()

	var err error
	var stmt *sqlx.NamedStmt
	query := fmt.Sprintf(queryInsertMessage, table)
	stmt, err = tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return Message{}, err
	}

	var rows *sql.Rows
	rows, err = stmt.QueryContext(ctx, map[string]interface{}{
		"id":              msg.Id,
		"sender_id":       msg.SenderId,
		"receiver_id":     msg.ReceiverId,
		"conversation_id": msg.ConversationId,
		"data":            msg.Data,
	})
	if err != nil {
		return Message{}, err
	}

	if rows.Next() {
		err = rows.Scan(&msg.Id)
		if err != nil {
			return Message{}, err
		}
	}

	return msg, nil
}

func (s *postgres) InsertMessage(c context.Context, msg Message, table string) (Message, error) {
	ctx, span := otel.Tracer("").Start(c, "repository.persistence.InsertReadMessage")
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
	query := fmt.Sprintf(queryInsertMessage, table)
	stmt, err = tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return Message{}, err
	}

	var rows *sql.Rows
	rows, err = stmt.QueryContext(ctx, map[string]interface{}{
		"id":              msg.Id,
		"sender_id":       msg.SenderId,
		"receiver_id":     msg.ReceiverId,
		"conversation_id": msg.ConversationId,
		"data":            msg.Data,
	})
	if err != nil {
		return Message{}, err
	}

	if rows.Next() {
		err = rows.Scan(&msg.Id)
		if err != nil {
			return Message{}, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return Message{}, err
	}
	return msg, nil
}

func (s *postgres) GetReadConversation(c context.Context, conversationId string, table string) ([]Message, error) {
	ctx, span := otel.Tracer("").Start(c, "repository.persistence.GetReadConversation")
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
	query := fmt.Sprintf(queryConversation, table)
	stmt, err = tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return []Message{}, err
	}

	var rows *sql.Rows
	rows, err = stmt.QueryContext(ctx, map[string]interface{}{
		"conversation_id": conversationId,
	})
	if err != nil {
		return []Message{}, err
	}

	messages := []Message{}
	for rows.Next() {
		m := Message{}
		scanErr := rows.Scan(&m.Id, &m.SenderId, &m.ReceiverId, &m.ConversationId, &m.Data, &m.CreateTime)
		if scanErr != nil {
			log.Println(scanErr.Error())
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

func (s *postgres) ReadMessage(c context.Context, conversationId string) error {
	ctx, span := otel.Tracer("").Start(c, "repository.persistence.DeleteUnreadMessage")
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
	stmt, err = tx.PrepareNamedContext(ctx, queryDeleteUnreadMessage)
	if err != nil {
		return err
	}

	var rows *sql.Rows
	rows, err = stmt.QueryContext(ctx, map[string]interface{}{
		"conversation_id": conversationId,
	})
	if err != nil {
		return err
	}

	for rows.Next() {
		m := Message{}
		rows.Scan(&m.Id, &m.SenderId, &m.ReceiverId, &m.ConversationId, &m.Data, &m.CreateTime)
		_, err = s.insertReadMessageByTx(ctx, tx, m, TableReadMessage)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
