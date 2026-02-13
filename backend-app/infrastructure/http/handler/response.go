package handler

import (
	"encoding/json"
	"net/http"
)

// Response is the standardized API response format
type Response struct {
	Data  any    `json:"data"`
	Error string `json:"error,omitempty"`
	Meta  Meta   `json:"meta"`
}

// Meta contains response metadata
type Meta struct {
	ResponseTime float64 `json:"responseTime"`
}

// JSON writes a standardized JSON response
func JSON(w http.ResponseWriter, status int, data any) error {
	resp := Response{
		Data: data,
		Meta: Meta{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(resp)
}

// JSONWithMeta writes a JSON response with custom metadata
func JSONWithMeta(w http.ResponseWriter, status int, data any, meta Meta) error {
	resp := Response{
		Data: data,
		Meta: meta,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(resp)
}

// Error writes an error response
func Error(w http.ResponseWriter, status int, message string) error {
	resp := Response{
		Error: message,
		Meta:  Meta{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(resp)
}

// OK writes a 200 OK response
func OK(w http.ResponseWriter, data any) error {
	return JSON(w, http.StatusOK, data)
}

// Created writes a 201 Created response
func Created(w http.ResponseWriter, data any) error {
	return JSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest writes a 400 error response
func BadRequest(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusBadRequest, message)
}

// Unauthorized writes a 401 error response
func Unauthorized(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusUnauthorized, message)
}

// Forbidden writes a 403 error response
func Forbidden(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusForbidden, message)
}

// NotFound writes a 404 error response
func NotFound(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusNotFound, message)
}

// Conflict writes a 409 error response
func Conflict(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusConflict, message)
}

// InternalError writes a 500 error response
func InternalError(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusInternalServerError, message)
}
