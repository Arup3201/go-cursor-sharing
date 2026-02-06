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

type Cursor struct {
	Id    string `json:"id"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Color string `json:"color"`
}

type Client struct {
	id    string
	conn  *websocket.Conn
	color string
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		if r.Header.Get("Origin") == "http://127.0.0.1:5500" {
			return true
		}

		return false
	},
}

var clients = map[string]Client{}
var subscribers = make(chan Cursor)

func publish(id string) {
	client, ok := clients[id]
	if !ok {
		log.Printf("[ERROR] client not found\n")
		return
	}

	for {
		mt, p, err := client.conn.ReadMessage()
		if err != nil {
			log.Printf("[ERROR] read message from publisher: %s\n", err)
			client.conn.Close()
			delete(clients, id)
			return
		}

		if mt == websocket.TextMessage {
			var c Cursor
			if err := json.Unmarshal(p, &c); err != nil {
				log.Printf("[ERROR] json unmarshal publisher data: %s", err)
				continue
			}

			cursor := Cursor{
				Id:    client.id,
				Color: client.color,
				X:     c.X,
				Y:     c.Y,
			}

			subscribers <- cursor
		}
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ERROR] upgrade: %s\n", err)
		return
	}

	id := uuid.NewString()

	randomColorInt := rand.Intn(0x1000000)
	hexColor := fmt.Sprintf("#%06X", randomColorInt)

	clients[id] = Client{
		id:    id,
		conn:  conn,
		color: hexColor,
	}
	log.Printf("[INFO] Registered client %s\n", id)

	go func() {
		publish(id)
	}()

	for {
		for cursor := range subscribers {

			// Don't send it to itself
			if cursor.Id == id {
				continue
			}

			data, err := json.Marshal(cursor)
			if err != nil {
				log.Printf("[ERROR] send data to subscriber marshal error: %s", err)
				continue
			}

			err = conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Printf("[ERROR] send data to subscriber error: %s\n", err)
				return
			}
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	http.HandleFunc("/ws", echo)
	err := http.ListenAndServe("127.0.0.1:8081", nil)
	if err != nil {
		log.Fatalf("[ERROR] service listen and serve: %s", err)
	}
}
