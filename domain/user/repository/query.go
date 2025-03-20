package repository

import "errors"

var (
	QueryInsertUser = "INSERT INTO users (username) VALUES (:username) RETURNING id"
	QueryGetUser    = "SELECT id, username FROM users WHERE username = :username"
)

var (
	ErrUserNotFound = errors.New("user not found")
)
