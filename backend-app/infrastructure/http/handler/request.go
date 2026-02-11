package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// GetUserID retrieves the user ID from request context
func GetUserIDFromContext(r *http.Request) string {
	if userID, ok := r.Context().Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// QueryParam returns a query parameter value
func QueryParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// PathVar returns a path variable using gorilla/mux
func PathVar(r *http.Request, name string) string {
	return mux.Vars(r)[name]
}

// BindJSON parses request body as JSON
func BindJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

// MustPathVar returns a path variable or writes a 400 error
func MustPathVar(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	value := PathVar(r, name)
	if value == "" {
		return "", BadRequest(w, fmt.Sprintf("path variable '%s' is required", name))
	}
	return value, nil
}

// MustQuery returns a query parameter or writes a 400 error
func MustQuery(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	value := QueryParam(r, name)
	if value == "" {
		return "", BadRequest(w, fmt.Sprintf("query parameter '%s' is required", name))
	}
	return value, nil
}
