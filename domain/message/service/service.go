package service

import (
	"context"
	"sort"

	"github.com/msyamsula/portofolio/domain/message/repository"
	"go.opentelemetry.io/otel"
)

type Service struct {
	Persistence PersistenceLayer
}

type Dependencies struct {
	Persistence PersistenceLayer
}

func New(dep Dependencies) *Service {
	return &Service{
		Persistence: dep.Persistence,
	}
}

type PersistenceLayer interface {
	AddMessage(c context.Context, msg repository.Message) (repository.Message, error)
	GetConversation(c context.Context, senderId, receiverId int64) ([]repository.Message, error)
}

func (s *Service) AddMessage(c context.Context, msg repository.Message) (repository.Message, error) {
	ctx, span := otel.Tracer("").Start(c, "service.AddMessage")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	// validation
	if msg.SenderId <= 0 ||
		msg.ReceiverId <= 0 ||
		msg.SenderId == msg.ReceiverId ||
		msg.Text == "" {
		err = ErrBadRequest
		return repository.Message{}, err
	}

	msg, err = s.Persistence.AddMessage(ctx, msg)
	if err != nil {
		return repository.Message{}, err
	}

	return msg, nil
}

func (s *Service) GetConversation(c context.Context, idA, idB int64) ([]repository.Message, error) {
	ctx, span := otel.Tracer("").Start(c, "service.GetConversation")
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
		}
	}()

	// validation
	if idA == idB {
		err = ErrBadRequest
		return []repository.Message{}, err
	}

	var msgsA, msgsB []repository.Message
	msgsA, err = s.Persistence.GetConversation(ctx, idA, idB)
	if err != nil {
		return []repository.Message{}, err
	}
	msgsB, err = s.Persistence.GetConversation(ctx, idB, idA)
	if err != nil {
		return []repository.Message{}, err
	}

	msgs := append(msgsA, msgsB...)

	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].CreateTime.Before(msgs[j].CreateTime)
	})

	return msgs, nil
}
