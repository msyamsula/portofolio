package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	userbinary "github.com/msyamsula/portofolio/domain/user/binary"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {

	// create server routes
	r := mux.NewRouter()

	// run user binary
	userbinary.Run(r)

	// cors option
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                              // Allow all origins (adjust for security)
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "HEAD"}, // Allow methods
		AllowedHeaders:   []string{"Content-Type"},                   // Allow headers
		AllowCredentials: true,                                       // Allows credentials (cookies, authorization headers)
	})
	corsHandler := c.Handler(r)

	// server handler
	http.Handle("/", otelhttp.NewHandler(corsHandler, "")) // use otelhttp for telemetry
	http.Handle("/metrics", promhttp.Handler())            // endpoint exporter, for prometheus scrapping

	// server start
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal("error in port format", err)
	}
	err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
