package repository

var (
	QueryInsertMessage   = "INSERT INTO messages (sender_id, receiver_id, text) VALUES (:sender_id, :receiver_id, :text) RETURNING id"
	QueryGetConversation = "SELECT id, sender_id, receiver_id, text, create_time, is_read FROM messages m WHERE sender_id = :sender_id AND receiver_id = :receiver_id"
	QueryReadMessage     = "UPDATE messages SET is_read = true WHERE sender_id = :sender_id AND receiver_id = :receiver_id"
	QueryUpdateUnread    = "insert into unread (sender_id, receiver_id, unread) values (:sender_id, :receiver_id, :unread) on conflict (sender_id, receiver_id) do update set unread = excluded.unread returning id;"
)
