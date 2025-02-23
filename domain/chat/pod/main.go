// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

import (
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/domain/chat/websocket"
)

var addr = flag.String("addr", ":8080", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func hubReceiver(c chan *websocket.Hub) {
	for {
		hub := <-c
		go hub.Run()
	}
}

var (
	hubMap  = make(map[string]*websocket.Hub)
	hubSync = sync.Mutex{}
)

func main() {
	flag.Parse()
	// hub := websocket.NewHub()
	// go hub.Run()
	roomChan := make(chan *websocket.Hub)
	go hubReceiver(roomChan)

	r := mux.NewRouter()

	r.HandleFunc("/", serveHome)
	r.HandleFunc("/ws/{room}", func(w http.ResponseWriter, r *http.Request) {
		pathVariables := mux.Vars(r)
		room, ok := pathVariables["room"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hubSync.Lock()
		if hubMap[room] == nil {
			hub := websocket.NewHub()
			hubMap[room] = hub
		}
		hubSync.Unlock()

		roomChan <- hubMap[room]
		websocket.ServeWs(hubMap[room], w, r)
	})

	http.Handle("/", r)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
