package websocket

// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	HubChan = make(chan *Hub)
)

func HubEvent() {
	for {
		hub := <-HubChan
		go hub.Run()
	}
}

func HubCleaner() {
	for {
		for hubName, h := range hubMap {
			fmt.Println(hubName, len(h.clients), h.IsEmpty())
			if h.IsEmpty() {
				hubSync.Lock()
				delete(hubMap, hubName)
				hubSync.Unlock()
				HubGauge.Dec()
			}
		}
		time.Sleep(2 * time.Minute)
	}
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// name
	name string

	// register error
	registerError chan error
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	query := r.URL.Query()
	username := query.Get("username")
	fmt.Println("username", username)
	if username == "" {
		fmt.Println("username empty")
		status := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "username is empty")
		conn.WriteMessage(websocket.CloseMessage, status)
		return
	}
	client := &Client{
		hub: hub, conn: conn,
		send:          make(chan []byte, 256),
		name:          username,
		registerError: make(chan error),
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()

	err = <-client.registerError
	if err != nil {
		fmt.Println("goes here")
		status := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "username is already taken")
		conn.WriteMessage(websocket.CloseMessage, status)
		return
	}

}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// make username distinct in a hub
	clientUsername map[string]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:      make(chan []byte),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make(map[*Client]bool),
		clientUsername: make(map[string]bool),
	}
}
func (h *Hub) IsEmpty() bool {
	return len(h.clients) == 0
}

var (
	hubMap  = make(map[string]*Hub)
	hubSync = sync.Mutex{}
)

func ConnectionHandler(w http.ResponseWriter, r *http.Request) {

	pathVariables := mux.Vars(r)
	hubName, ok := pathVariables["hub"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hubSync.Lock()
	if hubMap[hubName] == nil {
		hubMap[hubName] = NewHub()
		HubGauge.Inc()
	}
	hubSync.Unlock()

	HubChan <- hubMap[hubName]

	ServeWs(hubMap[hubName], w, r)
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			fmt.Println(h.clientUsername)
			// fmt.Println(h.cli)
			if h.clientUsername[client.name] == false {
				h.clients[client] = true
				h.clientUsername[client.name] = true
				UserGauge.Inc()
			} else {
				// reject the connection
				fmt.Println("duplicate error")
				client.registerError <- errors.New("duplicate username")
			}
			fmt.Println(h.clientUsername)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				UserGauge.Dec()
				delete(h.clients, client)
				delete(h.clientUsername, client.name)
				close(client.send)
			}
		case message := <-h.broadcast:
			MessageCounter.Inc()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					UserGauge.Dec()
					close(client.send)
					delete(h.clients, client)
					delete(h.clientUsername, client.name)
				}
			}
		}
	}
}

// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
