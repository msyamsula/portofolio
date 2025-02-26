// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/domain/chat/websocket"
)

var addr = flag.String("addr", ":8080", "http service address")

// func serveHome(w http.ResponseWriter, r *http.Request) {
// 	log.Println(r.URL)
// 	if r.URL.Path != "/" {
// 		http.Error(w, "Not found", http.StatusNotFound)
// 		return
// 	}
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
// 	http.ServeFile(w, r, "home.html")
// }

func hubEvent(c chan *websocket.Hub) {
	for {
		hub := <-c
		go hub.Run()
	}
}

var (
	hubMap  = make(map[string]*websocket.Hub)
	hubSync = sync.Mutex{}
)

func hubCleaner() {
	for {
		for hubName, h := range hubMap {
			if h.IsEmpty() {
				delete(hubMap, hubName)
			}
		}
		time.Sleep(5 * time.Minute)
	}
}

func main() {
	// look at this implementation for guidance in using websocket
	// https://github.com/gorilla/websocket?tab=readme-ov-file
	flag.Parse()
	hubChan := make(chan *websocket.Hub)
	go hubEvent(hubChan) // listen for hub creation
	go hubCleaner()      // periodically check if hub isEmpty or not

	r := mux.NewRouter()

	// r.HandleFunc("/", serveHome)
	apiPrefix := "/api/chat"
	r.HandleFunc(fmt.Sprintf("%s%s", apiPrefix, "/ws/{hub}"), func(w http.ResponseWriter, r *http.Request) {
		pathVariables := mux.Vars(r)
		hubName, ok := pathVariables["hub"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hubSync.Lock()
		if hubMap[hubName] == nil {
			hub := websocket.NewHub()
			hubMap[hubName] = hub
		}
		hubSync.Unlock()

		hubChan <- hubMap[hubName]
		websocket.ServeWs(hubMap[hubName], w, r)
	})

	http.Handle("/", r)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
