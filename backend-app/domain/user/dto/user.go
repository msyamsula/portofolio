package dto

// UserData represents user information from OAuth provider
type UserData struct {
	ID    string `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

// TokenResponse represents the response from token operations
type TokenResponse struct {
	Message string  `json:"message,omitempty"`
	Error   string  `json:"error,omitempty"`
	Token   string  `json:"token,omitempty"`
}

// ValidateTokenResponse represents the response from validating a token
type ValidateTokenResponse struct {
	Message string  `json:"message,omitempty"`
	Error   string  `json:"error,omitempty"`
	Data    UserData `json:"data,omitempty"`
}
