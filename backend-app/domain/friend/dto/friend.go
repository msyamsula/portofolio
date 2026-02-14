package dto

// User represents a user in the system
type User struct {
	Username string `json:"username,omitempty"`
	ID       int64  `json:"id,omitempty"`
	Online   bool   `json:"online,omitempty"`
	Unread   int64  `json:"unread,omitempty"`
}

// AddFriendRequest represents a request to add a friend
type AddFriendRequest struct {
	SmallID int64 `json:"small_id"`
	BigID   int64 `json:"big_id"`
}

// GetFriendsRequest represents a request to get friends for a user
type GetFriendsRequest struct {
	ID int64 `json:"id,omitempty"`
}

// GetFriendsResponse represents the response from getting friends
type GetFriendsResponse struct {
	Message string  `json:"message,omitempty"`
	Error   string  `json:"error,omitempty"`
	Data    []User  `json:"data,omitempty"`
}

// AddFriendResponse represents the response from adding a friend
type AddFriendResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
