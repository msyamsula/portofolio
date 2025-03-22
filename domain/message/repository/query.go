package repository

var (
	QueryInsertMessage   = "INSERT INTO messages (sender_id, receiver_id, text) VALUES (:sender_id, :receiver_id, :text) RETURNING id"
	QueryGetConversation = "SELECT id, sender_id, receiver_id, text, create_time FROM messages m WHERE sender_id = :sender_id AND receiver_id = :receiver_id"
)
