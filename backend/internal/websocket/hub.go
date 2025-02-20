// Package websocket provides WebSocket functionality for real-time communication.
package websocket

import (
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Client represents a WebSocket connection
type Client struct {
	Conn     *websocket.Conn
	UserID   primitive.ObjectID
	IsActive bool
}

// Message structure for WebSocket communication
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Hub maintains the active WebSocket connections
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	Register   chan *Client  // Capitalized for external access
	Unregister chan *Client  // Capitalized for external access
	mutex      sync.Mutex
}

// NewHub initializes and returns a new WebSocket Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run starts the Hub to manage clients and broadcast messages
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client connected. Total clients: %d", len(h.clients))

		case client := <-h.Unregister:
			h.mutex.Lock()
			if _, exists := h.clients[client]; exists {
				delete(h.clients, client)
				client.Conn.Close()
			}
			h.mutex.Unlock()
			log.Printf("Client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mutex.Lock()
			for client := range h.clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					log.Printf("Error broadcasting to client: %v", err)
					client.Conn.Close()
					delete(h.clients, client)  // Ensure removal to prevent memory leaks
				}
			}
			h.mutex.Unlock()
		}
	}
}

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(userID primitive.ObjectID, message Message) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for client := range h.clients {
		if client.UserID == userID {
			if err := client.Conn.WriteJSON(message); err != nil {
				log.Printf("Error broadcasting to user %s: %v", userID.Hex(), err)
				client.Conn.Close()
				delete(h.clients, client)
			}
		}
	}
}

// BroadcastToAll sends a message to all connected clients
func (h *Hub) BroadcastToAll(message Message) {
	h.broadcast <- message
}
