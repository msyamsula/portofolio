package service

import (
	"errors"

	"github.com/msyamsula/portofolio/backend-app/message/persistence"
)

var (
	ErrBadRequest = errors.New("bad request in message body")
)

type Config struct {
	Persistence persistence.Persistence
}
