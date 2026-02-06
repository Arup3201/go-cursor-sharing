package main

type Hub struct {
	register chan *Client

	unregister chan *Client

	clients map[string]*Client

	broadcast chan Cursor
}

func newHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		broadcast:  make(chan Cursor),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.id] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client.id]; ok {
				delete(h.clients, client.id)
				close(client.send)
			}
		case cursor := <-h.broadcast:
			for _, client := range h.clients {
				if client.id == cursor.Id {
					continue
				}

				select {
				case client.send <- cursor:
				default:
					close(client.send)
					delete(h.clients, client.id)
				}
			}
		}
	}
}
