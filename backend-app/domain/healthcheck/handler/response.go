package handler

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string  `json:"status"`
	Uptime float64 `json:"uptime_seconds"`
}
