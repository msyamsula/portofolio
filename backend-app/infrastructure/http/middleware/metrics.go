package middleware

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/metrics"
)

// MetricsMiddleware records HTTP request metrics (counter by path and status)
func MetricsMiddleware(instruments *metrics.Instruments) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if instruments == nil {
			return next
		}

		// Create the handler function
		handlerFunc := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			start := time.Now()

			// Wrap response writer to capture status code
			wrappedWriter := wrapResponseWriter(w)

			// Call next handler
			next.ServeHTTP(wrappedWriter, r)

			// Record metrics after request completes
			duration := time.Since(start).Seconds()
			status := wrappedWriter.statusCode
			method := r.Method
			path := r.URL.Path
			if route := mux.CurrentRoute(r); route != nil {
				if template, err := route.GetPathTemplate(); err == nil {
					path = template
				} else if regex, err := route.GetPathRegexp(); err == nil {
					path = regex
				}
			}

			instruments.RecordRequest(ctx, method, path, status, duration)
			instruments.SetResponseTime(duration)

			logger.Debug("request metrics recorded", map[string]any{
				"method":   method,
				"path":     path,
				"status":   status,
				"duration": duration,
			})
		}

		return http.HandlerFunc(handlerFunc)
	}
}

// wrapResponseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code when set
func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// StatusCode returns the captured status code
func (w *responseWriter) StatusCode() int {
	return w.statusCode
}

// Hijack implements http.Hijacker interface
func (w *responseWriter) Hijack() (c interface{}, rw interface{}, err error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// Flush implements http.Flusher interface
func (w *responseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// wrapResponseWriter wraps a response writer to capture status code
func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     200, // Default to 200 OK
	}
}
