package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

const sessionDuration = 3 * time.Hour

type Session struct {
	Token     *oauth2.Token
	Email     string
	Name      string
	CreatedAt time.Time
}

func (s *Session) IsExpired() bool {
	return time.Since(s.CreatedAt) > sessionDuration
}

// RefreshIfNeeded uses the refresh token to get a new access token if expired.
func (s *Session) RefreshIfNeeded(cfg *oauth2.Config) error {
	if s.Token.Valid() {
		return nil
	}
	// oauth2 TokenSource handles refresh automatically
	src := cfg.TokenSource(context.Background(), s.Token)
	newToken, err := src.Token()
	if err != nil {
		return err
	}
	s.Token = newToken
	return nil
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]*Session)}
}

func (s *SessionStore) Create(token *oauth2.Token, email, name string) string {
	id := generateID()
	s.mu.Lock()
	s.sessions[id] = &Session{
		Token:     token,
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
	}
	s.mu.Unlock()
	return id
}

func (s *SessionStore) Get(id string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess := s.sessions[id]
	if sess == nil {
		return nil
	}
	if sess.IsExpired() {
		return nil
	}
	return sess
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

func (s *SessionStore) GetFromRequest(r *http.Request) *Session {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil
	}
	return s.Get(cookie.Value)
}

func generateID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
