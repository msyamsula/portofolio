package middleware

import "net/http"

type middleware interface {
	AuthMiddleware(next http.Handler) http.Handler
}

func NewMiddleware(userHost string) middleware {
	return &userMiddleware{
		userHost: userHost,
	}
}
