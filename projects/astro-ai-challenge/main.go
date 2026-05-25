package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/people/v1"
)

var (
	oauthConfig  *oauth2.Config
	sessions     *SessionStore
	claudePath   string
)

func main() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}

	claudePath = os.Getenv("CLAUDE_PATH")
	if claudePath == "" {
		claudePath = "claude"
	}

	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes: []string{
			calendar.CalendarScope,
			people.ContactsReadonlyScope,
			people.DirectoryReadonlyScope,
			"openid", "email", "profile",
		},
		Endpoint:    google.Endpoint,
		RedirectURL: "http://localhost:8080/auth/callback",
	}

	sessions = NewSessionStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/auth/login", handleLogin)
	mux.HandleFunc("/auth/callback", handleCallback)
	mux.HandleFunc("/auth/status", handleAuthStatus)
	mux.HandleFunc("/auth/logout", handleLogout)
	mux.HandleFunc("/api/parse", handleParse)
	mux.HandleFunc("/api/resolve", handleResolve)
	mux.HandleFunc("/api/slots", handleSlots)
	mux.HandleFunc("/api/create", handleCreate)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Calendar AI running at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, mux))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateID()
	url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	email, name, err := getUserEmail(context.Background(), token, oauthConfig)
	if err != nil {
		http.Error(w, "failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sessionID := sessions.Create(token, email, name)
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionDuration.Seconds()),
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	session := sessions.GetFromRequest(r)
	w.Header().Set("Content-Type", "application/json")
	if session == nil {
		json.NewEncoder(w).Encode(map[string]any{"authenticated": false})
		return
	}
	json.NewEncoder(w).Encode(map[string]any{
		"authenticated": true,
		"email":         session.Email,
		"name":          session.Name,
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		sessions.Delete(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func requireSession(r *http.Request) (*Session, error) {
	session := sessions.GetFromRequest(r)
	if session == nil {
		return nil, fmt.Errorf("not authenticated")
	}
	// Auto-refresh token if expired (Google access tokens last ~1 hour)
	if err := session.RefreshIfNeeded(oauthConfig); err != nil {
		return nil, fmt.Errorf("session expired, please sign in again")
	}
	return session, nil
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := requireSession(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var body struct {
		Prompt   string         `json:"prompt"`
		Previous *ParsedMeeting `json:"previous,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	meeting, err := parseMeetingPrompt(claudePath, body.Prompt, body.Previous)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meeting)
}

func handleResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := requireSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var body struct {
		Names []string `json:"names"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results, err := resolveContacts(context.Background(), session.Token, oauthConfig, body.Names)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func handleSlots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := requireSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var body struct {
		Emails          []string `json:"emails"`
		DurationMinutes int      `json:"duration_minutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if body.DurationMinutes == 0 {
		body.DurationMinutes = 60
	}

	// Include the user's own calendar
	allEmails := append([]string{session.Email}, body.Emails...)

	slots, err := findAvailableSlots(context.Background(), session.Token, oauthConfig, allEmails, body.DurationMinutes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slots)
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := requireSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var body struct {
		Title           string   `json:"title"`
		AttendeeEmails  []string `json:"attendee_emails"`
		StartTime       string   `json:"start_time"`
		DurationMinutes int      `json:"duration_minutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	start, err := time.Parse(time.RFC3339, body.StartTime)
	if err != nil {
		start, err = time.Parse("2006-01-02T15:04:05", body.StartTime)
		if err != nil {
			http.Error(w, "invalid start_time format", http.StatusBadRequest)
			return
		}
		start = start.In(time.Now().Location())
	}

	// Enforce 1 month limit
	if start.After(time.Now().AddDate(0, 1, 0)) {
		http.Error(w, "cannot schedule more than 1 month ahead", http.StatusBadRequest)
		return
	}

	end := start.Add(time.Duration(body.DurationMinutes) * time.Minute)

	event, err := createCalendarEvent(context.Background(), session.Token, oauthConfig, body.Title, body.AttendeeEmails, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"event_id": event.Id,
		"link":     event.HtmlLink,
		"summary":  event.Summary,
		"start":    event.Start.DateTime,
		"end":      event.End.DateTime,
	})
}
