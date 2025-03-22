package service

import "errors"

var (
	ErrBadRequest = errors.New("bad request in message body")
)
