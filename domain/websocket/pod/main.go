// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/msyamsula/portofolio/domain/telemetry"
	"github.com/msyamsula/portofolio/domain/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	// look at this implementation for guidance in using websocket
	// https://github.com/gorilla/websocket?tab=readme-ov-file
	flag.Parse()

	godotenv.Load(".env")

	// instrumentation
	telemetry.InitializeTelemetryTracing("chat-server", os.Getenv("JAEGER_HOST"))

	// prometheus metrics
	prometheus.MustRegister(websocket.HubGauge)
	prometheus.MustRegister(websocket.UserGauge)
	prometheus.MustRegister(websocket.MessageCounter)

	// some background process related to hub
	go websocket.HubEvent()   // listen for hub creation
	go websocket.HubCleaner() // periodically check if hub isEmpty or not

	r := mux.NewRouter()

	apiPrefix := "/chat"
	r.HandleFunc(fmt.Sprintf("%s%s", apiPrefix, "/ws/{hub}"), websocket.ConnectionHandler)

	http.Handle("/", otelhttp.NewHandler(r, ""))
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
