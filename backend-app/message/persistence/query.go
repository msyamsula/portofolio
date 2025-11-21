package persistence

var (
	queryInsertMessage = "INSERT INTO %s (id, sender_id, receiver_id, conversation_id, data) VALUES (:id, :sender_id, :receiver_id, :conversation_id, :data) RETURNING id"
	queryConversation  = "SELECT id, sender_id, receiver_id, conversation_id, data, create_time FROM %s m WHERE conversation_id = :conversation_id ORDER BY create_time"
	queryDeleteMessage = "DELETE FROM %s WHERE conversation_id = :conversation_id RETURNING id, sender_id, receiver_id, conversation_id, data, create_time;"

	// queryConversation = "INSERT INTO read_messages (id, sender_id, receiver_id, conversation_id, data) VALUES (:id, :sender_id, :receiver_id, :conversation_id, :text) RETURNING id"
	// QueryReadMessage  = "UPDATE messages SET is_read = true WHERE sender_id = :sender_id AND receiver_id = :receiver_id"
	// QueryUpdateUnread = "insert into unread (sender_id, receiver_id, unread) values (:sender_id, :receiver_id, :unread) on conflict (sender_id, receiver_id) do update set unread = excluded.unread returning id;"
)
