package middleware

import (
	"net/http"

	"github.com/msyamsula/portofolio/backend-app/infrastructure/http/handler"
)

// XPortofolioMiddleware validates the x-portofolio header.
func XPortofolioMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-portofolio") == "" {
			_ = handler.Unauthorized(w, "missing x-portofolio header")
			return
		}
		next.ServeHTTP(w, r)
	})
}
