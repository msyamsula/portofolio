package service

import (
	"context"
	"database/sql"

	"github.com/msyamsula/portofolio/backend-app/message/persistence"
	"go.opentelemetry.io/otel"
)

type service struct {
	persistence persistence.Persistence
}

func (s *service) InsertUnreadMessage(c context.Context, msg persistence.Message) (persistence.Message, error) {
	ctx, span := otel.Tracer("").Start(c, "service.InsertUnreadMessage")
	defer span.End()

	var err error

	// validation
	if msg.SenderId <= 0 ||
		msg.ReceiverId <= 0 ||
		msg.SenderId == msg.ReceiverId ||
		msg.Data == "" {
		err = ErrBadRequest
		return persistence.Message{}, err
	}

	// open tx and defer rollback
	tx := s.persistence.MustBeginTx(c, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	msg, err = s.persistence.InsertMessage(ctx, tx, msg, persistence.TableMessage)
	if err != nil {
		return persistence.Message{}, err
	}

	err = tx.Commit()
	if err != nil {
		return persistence.Message{}, err
	}
	return msg, nil
}

func (s *service) GetConversation(c context.Context, conversationId string) ([]persistence.Message, error) {
	var err error

	// open tx and defer
	tx := s.persistence.MustBeginTx(c, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	conversations := []persistence.Message{}
	conversations, err = s.persistence.GetMessage(c, tx, conversationId, persistence.TableMessage)
	if err != nil {
		return []persistence.Message{}, err
	}

	err = tx.Commit()
	if err != nil {
		return []persistence.Message{}, err
	}

	return conversations, nil
}
