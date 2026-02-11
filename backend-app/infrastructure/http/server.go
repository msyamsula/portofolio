package minimal

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

// Server wraps gorilla/mux with minimal boilerplate
type Server struct {
	router *mux.Router
	server *http.Server
}

// New creates a new server using gorilla/mux
func New(addr string) *Server {
	r := mux.NewRouter()
	return &Server{
		router: r,
		server: &http.Server{
			Addr:         addr,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Router returns the gorilla/mux router for registering routes
func (s *Server) Router() *mux.Router {
	return s.router
}

// Start starts the server
func (s *Server) Start() error {
	s.server.Handler = s.router

	fmt.Printf("Starting server on %s\n", s.server.Addr)

	errChan := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		fmt.Println("\nShutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
}
