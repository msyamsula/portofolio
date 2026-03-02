package middleware

import (
"context"
"encoding/json"
"net/http"
"net/http/httptest"
"testing"

"github.com/stretchr/testify/suite"
)

type MiddlewareTestSuite struct {
suite.Suite
}

// --- Chain ---

func (s *MiddlewareTestSuite) TestChain_AppliesInOrder() {
var order []string

mw1 := func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
order = append(order, "mw1-before")
next.ServeHTTP(w, r)
order = append(order, "mw1-after")
})
}
mw2 := func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
order = append(order, "mw2-before")
next.ServeHTTP(w, r)
order = append(order, "mw2-after")
})
}

chained := Chain(mw1, mw2)
handler := chained(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
order = append(order, "handler")
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal([]string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}, order)
}

func (s *MiddlewareTestSuite) TestChain_Empty() {
called := false
chained := Chain()
handler := chained(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
called = true
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.True(called)
}

// --- RecoveryMiddleware ---

func (s *MiddlewareTestSuite) TestRecoveryMiddleware_NoPanic() {
handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusOK, rec.Code)
}

func (s *MiddlewareTestSuite) TestRecoveryMiddleware_RecoversPanic() {
handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
panic("test panic")
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()

s.NotPanics(func() {
handler.ServeHTTP(rec, req)
})
s.Equal(http.StatusInternalServerError, rec.Code)
}

// --- LoggingMiddleware ---

func (s *MiddlewareTestSuite) TestLoggingMiddleware_PassesThrough() {
called := false
handler := LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
called = true
w.WriteHeader(http.StatusOK)
}))

req := httptest.NewRequest(http.MethodGet, "/test?q=1", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.True(called)
s.Equal(http.StatusOK, rec.Code)
}

// --- CORSMiddleware ---

func (s *MiddlewareTestSuite) TestCORSMiddleware_AllowedOrigin() {
handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
req.Header.Set("Origin", "http://localhost:5500")
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal("http://localhost:5500", rec.Header().Get("Access-Control-Allow-Origin"))
s.Equal("Origin", rec.Header().Get("Vary"))
s.Equal(http.StatusOK, rec.Code)
}

func (s *MiddlewareTestSuite) TestCORSMiddleware_DisallowedOrigin() {
handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
req.Header.Set("Origin", "http://evil.com")
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal("", rec.Header().Get("Access-Control-Allow-Origin"))
s.Equal(http.StatusOK, rec.Code)
}

func (s *MiddlewareTestSuite) TestCORSMiddleware_OptionsRequest() {
handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusTeapot)
}))

req := httptest.NewRequest(http.MethodOptions, "/", nil)
req.Header.Set("Origin", "http://localhost:5500")
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusOK, rec.Code)
s.Contains(rec.Header().Get("Access-Control-Allow-Methods"), "GET")
s.Contains(rec.Header().Get("Access-Control-Allow-Methods"), "POST")
}

func (s *MiddlewareTestSuite) TestCORSMiddleware_NoOrigin() {
handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal("", rec.Header().Get("Access-Control-Allow-Origin"))
}

// --- AuthMiddleware ---

func (s *MiddlewareTestSuite) TestAuthMiddleware_ValidBearer() {
handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
userID, ok := r.Context().Value("user_id").(string)
s.True(ok)
s.Equal("my-token-123", userID)
w.WriteHeader(http.StatusOK)
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
req.Header.Set("Authorization", "Bearer my-token-123")
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusOK, rec.Code)
}

func (s *MiddlewareTestSuite) TestAuthMiddleware_MissingHeader() {
handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
s.Fail("should not reach handler")
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *MiddlewareTestSuite) TestAuthMiddleware_InvalidFormat() {
handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
s.Fail("should not reach handler")
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
req.Header.Set("Authorization", "Basic abc123")
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusUnauthorized, rec.Code)
}

// --- AdminMiddleware ---

func (s *MiddlewareTestSuite) TestAdminMiddleware_WithUserID() {
handler := AdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
ctx := context.WithValue(req.Context(), "user_id", "admin-user")
req = req.WithContext(ctx)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusOK, rec.Code)
}

func (s *MiddlewareTestSuite) TestAdminMiddleware_NoUserID() {
handler := AdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
s.Fail("should not reach handler")
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusUnauthorized, rec.Code)
}

// --- ContentTypeMiddleware ---

func (s *MiddlewareTestSuite) TestContentTypeMiddleware_POSTWithJSON() {
handler := ContentTypeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
}))

req := httptest.NewRequest(http.MethodPost, "/", nil)
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusOK, rec.Code)
}

func (s *MiddlewareTestSuite) TestContentTypeMiddleware_POSTWithoutJSON() {
handler := ContentTypeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
s.Fail("should not reach handler")
}))

req := httptest.NewRequest(http.MethodPost, "/", nil)
req.Header.Set("Content-Type", "text/plain")
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *MiddlewareTestSuite) TestContentTypeMiddleware_PUTWithoutJSON() {
handler := ContentTypeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
s.Fail("should not reach handler")
}))

req := httptest.NewRequest(http.MethodPut, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *MiddlewareTestSuite) TestContentTypeMiddleware_GETPassesThrough() {
called := false
handler := ContentTypeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
called = true
w.WriteHeader(http.StatusOK)
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.True(called)
s.Equal(http.StatusOK, rec.Code)
}

// --- XPortofolioMiddleware ---

func (s *MiddlewareTestSuite) TestXPortofolioMiddleware_WithHeader() {
handler := XPortofolioMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
req.Header.Set("x-portofolio", "true")
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusOK, rec.Code)
}

func (s *MiddlewareTestSuite) TestXPortofolioMiddleware_WithoutHeader() {
handler := XPortofolioMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
s.Fail("should not reach handler")
}))

req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusUnauthorized, rec.Code)

var resp map[string]any
_ = json.NewDecoder(rec.Body).Decode(&resp)
s.Contains(resp["error"], "x-portofolio")
}

// --- ResponseTimeMiddleware ---

func (s *MiddlewareTestSuite) TestResponseTimeMiddleware_InjectsResponseTime() {
inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(map[string]any{
"data": "hello",
"meta": map[string]any{},
})
})

handler := ResponseTimeMiddleware(inner)
req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusOK, rec.Code)
s.Contains(rec.Header().Get("Content-Type"), "application/json")

var body map[string]any
err := json.NewDecoder(rec.Body).Decode(&body)
s.NoError(err)

meta, ok := body["meta"].(map[string]any)
s.True(ok)
rt, ok := meta["responseTime"].(float64)
s.True(ok)
s.Greater(rt, float64(0))
}

func (s *MiddlewareTestSuite) TestResponseTimeMiddleware_NonJSONPassesThrough() {
inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "text/plain")
w.WriteHeader(http.StatusOK)
w.Write([]byte("hello world"))
})

handler := ResponseTimeMiddleware(inner)
req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusOK, rec.Code)
s.Equal("hello world", rec.Body.String())
}

func (s *MiddlewareTestSuite) TestResponseTimeMiddleware_NoBody() {
inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusNoContent)
})

handler := ResponseTimeMiddleware(inner)
req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusNoContent, rec.Code)
}

func (s *MiddlewareTestSuite) TestResponseTimeMiddleware_NoMetaKey() {
inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(map[string]any{"data": "hello"})
})

handler := ResponseTimeMiddleware(inner)
req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

var body map[string]any
json.NewDecoder(rec.Body).Decode(&body)
_, hasMeta := body["meta"]
s.False(hasMeta)
}

func (s *MiddlewareTestSuite) TestResponseTimeMiddleware_InvalidJSON() {
inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
w.Write([]byte("not valid json{"))
})

handler := ResponseTimeMiddleware(inner)
req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.Equal(http.StatusOK, rec.Code)
s.Equal("not valid json{", rec.Body.String())
}

// --- MetricsMiddleware ---

func (s *MiddlewareTestSuite) TestMetricsMiddleware_NilInstruments_PassesThrough() {
called := false
inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
called = true
w.WriteHeader(http.StatusOK)
})

handler := MetricsMiddleware(nil)(inner)
req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

s.True(called)
s.Equal(http.StatusOK, rec.Code)
}

// --- responseWriter wrapper ---

func (s *MiddlewareTestSuite) TestWrapResponseWriter_DefaultStatusCode() {
rec := httptest.NewRecorder()
wrapped := wrapResponseWriter(rec)
s.Equal(200, wrapped.statusCode)
s.Equal(200, wrapped.StatusCode())
}

func (s *MiddlewareTestSuite) TestWrapResponseWriter_CapturesStatusCode() {
rec := httptest.NewRecorder()
wrapped := wrapResponseWriter(rec)
wrapped.WriteHeader(http.StatusNotFound)
s.Equal(http.StatusNotFound, wrapped.StatusCode())
}

func (s *MiddlewareTestSuite) TestWrapResponseWriter_Flush() {
rec := httptest.NewRecorder()
wrapped := wrapResponseWriter(rec)
s.NotPanics(func() {
wrapped.Flush()
})
}

// --- bufferedResponseWriter ---

func (s *MiddlewareTestSuite) TestBufferedResponseWriter_Header() {
bw := newBufferedResponseWriter()
bw.Header().Set("X-Test", "value")
s.Equal("value", bw.Header().Get("X-Test"))
}

func (s *MiddlewareTestSuite) TestBufferedResponseWriter_WriteHeader() {
bw := newBufferedResponseWriter()
bw.WriteHeader(http.StatusCreated)
s.Equal(http.StatusCreated, bw.statusCode)
}

func (s *MiddlewareTestSuite) TestBufferedResponseWriter_Write() {
bw := newBufferedResponseWriter()
n, err := bw.Write([]byte("hello"))
s.NoError(err)
s.Equal(5, n)
s.Equal("hello", bw.body.String())
}

func TestMiddlewareSuite(t *testing.T) {
suite.Run(t, new(MiddlewareTestSuite))
}
