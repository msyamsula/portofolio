package handler

// ShortenRequest represents a URL shortening request
type ShortenRequest struct {
	LongURL string `json:"long_url"`
}
