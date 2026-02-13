package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// ResponseTimeMiddleware measures full request duration and injects it into response meta.
func ResponseTimeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		buffered := newBufferedResponseWriter()
		next.ServeHTTP(buffered, r)

		durationMs := float64(time.Since(start).Microseconds()) / 1000

		if buffered.body.Len() > 0 {
			contentType := buffered.header.Get("Content-Type")
			if contentType == "" || strings.Contains(contentType, "application/json") {
				updatedBody, updated := injectResponseTime(buffered.body.Bytes(), durationMs)
				if updated {
					buffered.body.Reset()
					buffered.body.Write(updatedBody)
					if contentType == "" {
						buffered.header.Set("Content-Type", "application/json")
					}
				}
			}
		}

		writeBufferedResponse(w, buffered)
	})
}

type bufferedResponseWriter struct {
	header     http.Header
	statusCode int
	body       *bytes.Buffer
}

func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{
		header:     make(http.Header),
		statusCode: http.StatusOK,
		body:       &bytes.Buffer{},
	}
}

func (w *bufferedResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *bufferedResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func writeBufferedResponse(w http.ResponseWriter, buffered *bufferedResponseWriter) {
	for key, values := range buffered.header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(buffered.statusCode)
	if buffered.body.Len() > 0 {
		_, _ = w.Write(buffered.body.Bytes())
	}
}

func injectResponseTime(body []byte, durationMs float64) ([]byte, bool) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return body, false
	}

	metaValue, ok := payload["meta"]
	if !ok {
		return body, false
	}

	meta, ok := metaValue.(map[string]any)
	if !ok {
		return body, false
	}

	meta["responseTime"] = durationMs
	payload["meta"] = meta

	updated, err := json.Marshal(payload)
	if err != nil {
		return body, false
	}

	return updated, true
}
