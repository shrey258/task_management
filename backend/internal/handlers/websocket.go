// Package handlers provides HTTP and WebSocket request handlers for the task management system.
package handlers

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/shrey258/task_management/internal/middleware"
	ws "github.com/shrey258/task_management/internal/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebSocketHandler manages WebSocket connections and message broadcasting.
type WebSocketHandler struct {
	hub *ws.Hub
}

// NewWebSocketHandler creates a new WebSocketHandler with the provided hub.
func NewWebSocketHandler(hub *ws.Hub) *WebSocketHandler {
	if hub == nil {
		panic("websocket hub cannot be nil")
	}
	return &WebSocketHandler{
		hub: hub,
	}
}

// HandleWebSocket handles individual WebSocket connections.
// It manages the lifecycle of the connection, including registration,
// message handling, and cleanup.
func (h *WebSocketHandler) HandleWebSocket(c *websocket.Conn) {
	if c == nil {
		log.Println("websocket: received nil connection")
		return
	}

	userID, ok := c.Locals("user_id").(primitive.ObjectID)
	if !ok {
		log.Println("websocket: invalid or missing user_id")
		c.Close()
		return
	}

	client := &ws.Client{
		Conn:     c,
		UserID:   userID,
		IsActive: true,
	}

	// Register client and ensure cleanup
	h.hub.Register <- client
	defer func() {
		h.hub.Unregister <- client
		client.IsActive = false
		c.Close()
	}()

	// Start message handling loop
	for {
		var message ws.Message
		if err := c.ReadJSON(&message); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		// Broadcast the message to all clients using the exported method
		h.hub.BroadcastToAll(message)
	}
}

// UpgradeConnection upgrades an HTTP connection to a WebSocket connection.
// It checks if the client requested a WebSocket upgrade and handles the upgrade process.
func (h *WebSocketHandler) UpgradeConnection(c *fiber.Ctx) error {
	// Check for auth token in query parameter
	token := c.Query("token")
	if token == "" {
		// Try getting token from Authorization header
		auth := c.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}
	}

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized: missing token",
		})
	}

	// Validate token and set user_id in context
	userID, err := middleware.ValidateToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized: invalid token",
		})
	}

	// Store user_id in locals for WebSocket handler
	c.Locals("user_id", userID)

	// Upgrade the connection
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return websocket.New(h.HandleWebSocket)(c)
	}

	return fiber.ErrUpgradeRequired
}
