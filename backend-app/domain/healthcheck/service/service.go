package service

import (
	"context"
	"time"
)

// Service defines the interface for health check operations
type Service interface {
	Check(ctx context.Context) (string, error)
	Uptime() float64
}

// service handles health check operations
type service struct {
	startTime time.Time
}

// New creates a new health check service
func New() Service {
	return &service{
		startTime: time.Now(),
	}
}

// Check returns the health status of the service
func (s *service) Check(ctx context.Context) (string, error) {
	// Simple health check - always return healthy
	// In a real implementation, you might check:
	// - Database connectivity
	// - Redis connectivity
	// - External service dependencies
	// - Disk space, memory usage, etc.
	return "healthy", nil
}

// Uptime returns the uptime of the service in seconds
func (s *service) Uptime() float64 {
	return time.Since(s.startTime).Seconds()
}
