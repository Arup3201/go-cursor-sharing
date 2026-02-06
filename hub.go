package main

type Hub struct {
	register chan *Client

	unregister chan *Client

	clients map[string]*Client

	broadcast chan Message
}

func newHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		broadcast:  make(chan Message),
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

				message := Message{
					Type: "remove",
					Cursor: &Cursor{
						Id:    client.id,
						Color: client.color,
						X:     client.lastX,
						Y:     client.lastY,
					},
				}

				for _, client := range h.clients {
					select {
					case client.send <- message:
					default: // don't block for slow clients
					}
				}
			}
		case message := <-h.broadcast:
			for _, client := range h.clients {
				if client.id == message.Cursor.Id {
					continue
				}

				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client.id)
				}
			}
		}
	}
}
