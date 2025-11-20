package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type consumer struct {
	server *http.Server
}

func (c *consumer) Start(ch chan os.Signal) {
	// create server routes
	r := mux.NewRouter()
	r.HandleFunc("/sns", func(w http.ResponseWriter, r *http.Request) {
		bInput, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("failed to read request body:", err)
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}

		log.Println("Received SNS message:", string(bInput))

		type message struct {
			Text string `json:"text"`
		}

		msg := message{
			Text: string(bInput),
		}

		json.NewEncoder(w).Encode(msg)
	}).Methods(http.MethodPost)

	tracedHandler := otelhttp.NewHandler(r, "")

	// cors option
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                              // Allow all origins (adjust for security)
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "HEAD"}, // Allow methods
		AllowedHeaders:   []string{"Content-Type"},                   // Allow headers
		AllowCredentials: true,                                       // Allows credentials (cookies, authorization headers)
	})
	corsHandler := cors.Handler(tracedHandler)

	port := "10000"
	c.server = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", port),
		Handler: corsHandler,
	}

	if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}

}

func (c *consumer) Stop() error {
	return c.server.Shutdown(context.Background())
}
