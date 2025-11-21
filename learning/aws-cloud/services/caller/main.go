package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Response struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type doFunction func(r *http.Request) (*http.Response, error)

func doWithRetryAndBackoff(do doFunction, r *http.Request, retry int) (*http.Response, error) {
	fmt.Println(r.URL.String())
	var lastErr error
	for i := 0; i < retry; i++ {
		fmt.Println("retry", i+1)
		resp, err := do(r)
		if err == nil {
			// return early
			return resp, nil
		}

		// retry block
		lastErr = err
		// exponential backoff
		time.Sleep(time.Duration(i*1) * 500 * time.Millisecond)
	}

	return &http.Response{}, lastErr
}

func call(ctx context.Context) (Response, error) {
	c := http.Client{
		Transport: http.DefaultTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			panic("TODO")
		},
		Jar:     nil,
		Timeout: 5 * time.Second,
	}

	host := os.Getenv("CALLEE_HOST")

	url := fmt.Sprintf("%s/ping", host)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Response{}, err
	}

	// retry and backoff
	resp, err := doWithRetryAndBackoff(c.Do, req, 3)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	r := Response{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return Response{}, err
	}

	return r, nil

}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var resp Response
	var err error
	defer func() {
		json.NewEncoder(w).Encode(resp)
	}()
	resp, err = call(ctx)
	if err != nil {
		resp.Error = err.Error()
		resp.Code = http.StatusInternalServerError
		return
	}

	resp.Code = http.StatusOK
}

func main() {
	// create server routes
	r := mux.NewRouter()

	// cors option
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                              // Allow all origins (adjust for security)
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "HEAD"}, // Allow methods
		AllowedHeaders:   []string{"Content-Type"},                   // Allow headers
		AllowCredentials: true,                                       // Allows credentials (cookies, authorization headers)
	})
	corsHandler := c.Handler(r)

	r.HandleFunc("/ping", handler).Methods(http.MethodGet)

	// dummy server
	http.Handle("/", otelhttp.NewHandler(corsHandler, "")) // use otelhttp for telemetry

	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", 8000), nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
