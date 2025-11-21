package repository

import "errors"

var (
	QueryInsertUser = "INSERT INTO users (username, online) VALUES (:username, :online) ON CONFLICT (username) DO UPDATE SET online = EXCLUDED.online RETURNING id"
	QueryGetUser    = "SELECT id, username, online FROM users WHERE username = :username"

	QueryAddFriend  = "INSERT INTO friendship (small_id, big_id) VALUES (:small_id, :big_id) RETURNING id"
	QueryGetFriends = `
	select l.id, l.username, l.online, coalesce(ur.unread,0) FROM
	(
		SELECT u.id, u.username, u.online FROM
			(SELECT big_id FROM friendship WHERE small_id = :user_id) f
		JOIN users u ON u.id = f.big_id -- half
		UNION ALL
		SELECT u.id, u.username, u.online FROM
			(SELECT small_id FROM friendship WHERE big_id = :user_id) f
		JOIN users u ON u.id = f.small_id -- other half
	) l left join unread ur on l.id = ur.sender_id and ur.receiver_id = :user_id
	`
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrIdMustBeDifferent = errors.New("id must be different")
)
