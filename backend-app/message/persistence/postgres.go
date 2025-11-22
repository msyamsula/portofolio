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

func (s *postgres) DeleteBulkMessage(ctx context.Context, tx *sqlx.Tx, conversationId string, table string) ([]Message, error) {
	var stmt *sqlx.NamedStmt
	var err error
	stmt, err = tx.PrepareNamedContext(ctx, fmt.Sprintf(queryDeleteMessage, table))
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

	deletedMessages := []Message{}
	for rows.Next() {
		m := Message{}
		err = rows.Scan(&m.Id, &m.SenderId, &m.ReceiverId, &m.ConversationId, &m.Data, &m.CreateTime)
		if err != nil {
			return []Message{}, err
		}
		deletedMessages = append(deletedMessages, m)
	}

	return deletedMessages, nil
}

func (s *postgres) InsertBulkMessage(c context.Context, tx *sqlx.Tx, msg []Message, table string) error {
	ctx, span := otel.Tracer("").Start(c, "repository.persistence.InsertBulkMessage")
	defer span.End()

	var err error
	for _, m := range msg {
		var stmt *sqlx.NamedStmt
		query := fmt.Sprintf(queryInsertMessage, TableMessage)
		stmt, err = tx.PrepareNamedContext(ctx, query)
		if err != nil {
			return err
		}

		var rows *sql.Rows
		rows, err = stmt.QueryContext(ctx, map[string]interface{}{
			"id":              m.Id,
			"sender_id":       m.SenderId,
			"receiver_id":     m.ReceiverId,
			"conversation_id": m.ConversationId,
			"data":            m.Data,
		})
		if err != nil {
			return err
		}

		if rows.Next() {
			err = rows.Scan(&m.Id)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *postgres) InsertMessage(c context.Context, tx *sqlx.Tx, msg Message, table string) (Message, error) {
	ctx, span := otel.Tracer("").Start(c, "repository.persistence.InsertMessage")
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

func (s *postgres) GetMessage(c context.Context, tx *sqlx.Tx, conversationId string, table string) ([]Message, error) {
	ctx, span := otel.Tracer("").Start(c, "repository.persistence.GetConversation")
	defer span.End()

	var err error
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

func (s *postgres) MustBeginTx(c context.Context, opt *sql.TxOptions) *sqlx.Tx {
	return s.MustBeginTx(c, opt)
}
