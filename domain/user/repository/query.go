package repository

import "errors"

var (
	QueryInsertUser = "INSERT INTO users (username) VALUES (:username) RETURNING id"
	QueryGetUser    = "SELECT id, username FROM users WHERE username = :username"

	QueryAddFriend  = "INSERT INTO friendship (small_id, big_id) VALUES (:small_id, :big_id) RETURNING id"
	QueryGetFriends = `SELECT * FROM
	(
		SELECT u.id, u.username FROM
			(SELECT big_id FROM friendship WHERE small_id = :user_id) f
		JOIN users u ON u.id = f.big_id -- half
		UNION ALL
		SELECT u.id, u.username FROM
			(SELECT small_id FROM friendship WHERE big_id = :user_id) f
		JOIN users u ON u.id = f.small_id -- other half
	) l;`
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrIdMustBeDifferent = errors.New("id must be different")
)
