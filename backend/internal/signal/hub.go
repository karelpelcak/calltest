package signal

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client

	rooms map[string]map[*Client]bool // roomID → klienti
	lock  sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rooms:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.lock.Lock()
			h.clients[client] = true
			if h.rooms[client.roomID] == nil {
				h.rooms[client.roomID] = make(map[*Client]bool)
			}
			h.rooms[client.roomID][client] = true
			h.lock.Unlock()

			log.Printf("✅ User '%s' connected to room '%s'", client.userID, client.roomID)

		case client := <-h.unregister:
			h.lock.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if h.rooms[client.roomID] != nil {
					delete(h.rooms[client.roomID], client)
					if len(h.rooms[client.roomID]) == 0 {
						delete(h.rooms, client.roomID)
					}
				}
				close(client.send)
				log.Printf("❌ User '%s' disconnected from room '%s'", client.userID, client.roomID)
			}
			h.lock.Unlock()

		case message := <-h.broadcast:
			h.lock.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.lock.Unlock()
		}
	}
}

// ----------------- Client -----------------

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	roomID string
	userID string
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	userID := r.URL.Query().Get("user")

	if roomID == "" || userID == "" {
		http.Error(w, "room and user query params are required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		roomID: roomID,
		userID: userID,
	}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		// broadcast jen klientům ve stejné room
		c.hub.lock.Lock()
		for cl := range c.hub.rooms[c.roomID] {
			select {
			case cl.send <- message:
			default:
				close(cl.send)
				delete(c.hub.clients, cl)
				delete(c.hub.rooms[c.roomID], cl)
			}
		}
		c.hub.lock.Unlock()
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}
