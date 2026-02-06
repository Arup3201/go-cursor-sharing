package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	WRITE_WAIT    = 10 * time.Second
	PONG_INTERVAL = 60 * time.Second
	PING_INTERVAL = PONG_INTERVAL * 9 / 10
	MESSAGE_LIMIT = 255
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		if r.Header.Get("Origin") == "http://127.0.0.1:5500" {
			return true
		}

		return false
	},
}

type Cursor struct {
	Id    string `json:"id"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Color string `json:"color"`
}

type Client struct {
	id    string
	color string

	conn *websocket.Conn

	hub *Hub

	send chan Cursor
}

// readPump pumps messages from client websocket to hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(MESSAGE_LIMIT)
	c.conn.SetReadDeadline(time.Now().Add(PONG_INTERVAL))
	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(PONG_INTERVAL))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("[ERROR] conn read message: %s\n", err)
			break
		}

		var cursor Cursor
		if err := json.Unmarshal(message, &cursor); err != nil {
			log.Printf("[ERROR] json unmarshal message to cursor: %s\n", err)
			continue
		}

		cursor.Id = c.id
		cursor.Color = c.color

		c.hub.broadcast <- cursor
	}
}

// writePump pumps messages from the hub to the client websocket
func (c *Client) writePump() {
	ticker := time.NewTicker(PING_INTERVAL)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case cursor, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))

			if !ok {
				// hub closed the connection
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("[ERROR] connection next writer: %s", err)
				return
			}

			message, err := json.Marshal(cursor)
			if err != nil {
				log.Printf("[ERROR] json marshal cursor to message: %s\n", err)
				continue
			}

			w.Write(message)

			if err = w.Close(); err != nil {
				log.Printf("[ERROR] write close: %s", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ERROR] upgrade: %s\n", err)
		return
	}

	id := uuid.NewString()

	randomColorInt := rand.Intn(0x1000000)
	hexColor := fmt.Sprintf("#%06X", randomColorInt)

	client := &Client{
		id:    id,
		color: hexColor,
		conn:  conn,
		hub:   hub,
		send:  make(chan Cursor),
	}

	hub.register <- client
	log.Printf("[INFO] Registered client %s\n", id)

	go client.readPump()
	go client.writePump()
}
