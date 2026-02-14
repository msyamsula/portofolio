package middleware

import (
	"net/http"
	"time"

	"github.com/msyamsula/portofolio/backend-app/pkg/logger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type userMiddleware struct {
	userHost string
}

func (mw *userMiddleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Implement authentication logic here
		// For example, check for an API key in headers

		// If authentication fails
		// http.Error(w, "Unauthorized", http.StatusUnauthorized)
		// return

		// If authentication succeeds
		if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}
		c := http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   1000 * time.Millisecond,
		}
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, mw.userHost, nil)
		if err != nil {
			logger.Logger.Errorf("auth middleware failed to create request %s", err.Error())
			http.Error(w, "auth middleware failed to create request", http.StatusInternalServerError)
			return
		}

		req.Header.Set("Authorization", r.Header.Get("Authorization"))

		resp, err := c.Do(req)
		if err != nil {
			logger.Logger.Errorf("user request failed %s", err.Error())
			http.Error(w, "user request failed", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Logger.Errorf("unauthorized user, status code %d", resp.StatusCode)
			http.Error(w, "unauthorized user", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
