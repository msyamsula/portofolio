package persistent

import "errors"

var (
	QueryInsertUser = "INSERT INTO users (username, online) VALUES (:username, :online) ON CONFLICT (username) DO UPDATE SET online = EXCLUDED.online RETURNING id"
	QueryGetUser    = "SELECT id, username, online FROM users WHERE username = :username"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrIdMustBeDifferent = errors.New("id must be different")
)
