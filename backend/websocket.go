package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// ========================================
// WEBSOCKET MANAGER
// ========================================

// WebSocketManager manages all WebSocket connections
type WebSocketManager struct {
	Clients   map[*websocket.Conn]bool
	Broadcast  chan []byte
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
}

// NewWebSocketManager creates new WebSocket manager
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		Clients:   make(map[*websocket.Conn]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}
}

// HandleWebSocket handles WebSocket connection
func (m *WebSocketManager) HandleWebSocket() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Upgrade HTTP to WebSocket
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins (in production, validate origin)
			},
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}

		// Register client
		m.Register <- conn

		// Handle incoming messages
		go m.handleConnection(conn)
	}
}

func (m *WebSocketManager) handleConnection(conn *websocket.Conn) {
	defer func() {
		m.Unregister <- conn
		conn.Close()
	}()

	// Read messages
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if messageType == websocket.TextMessage {
			var message map[string]interface{}
			json.Unmarshal(p, &message)

			// Handle client requests
			if message["type"] == "subscribe" {
				// Client wants to subscribe to a room
				// In production, implement room logic
				log.Printf("Client subscribed to: %v", message["room"])
			}
		}
	}
}

// Run starts the WebSocket manager
func (m *WebSocketManager) Run() {
	for {
		select {
		case client := <-m.Register:
			m.Clients[client] = true
			log.Println("Client connected. Total clients:", len(m.Clients))

		case client := <-m.Unregister:
			delete(m.Clients, client)
			log.Println("Client disconnected. Total clients:", len(m.Clients))

		case message := <-m.Broadcast:
			// Send message to all clients
			for client := range m.Clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Println("Error sending message to client:", err)
					delete(m.Clients, client)
				}
			}
		}
	}
}

// Broadcast sends message to all clients
func (m *WebSocketManager) Broadcast(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	m.Broadcast <- data
	return nil
}

// SendToClient sends message to specific client
func (m *WebSocketManager) SendToClient(client *websocket.Conn, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return client.WriteMessage(websocket.TextMessage, data)
}

// SendToRoom sends message to clients in a room (simplified - in production, use room management)
func (m *WebSocketManager) SendToRoom(room string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Send to all clients for now (in production, implement room filtering)
	for client := range m.Clients {
		err := client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			delete(m.Clients, client)
		}
	}

	return nil
}

// ========================================
// NOTIFICATION TYPES
// ========================================

// Notification represents a notification message
type Notification struct {
	Type      string                 `json:"type"`
	Action    string                 `json:"action"`
	Data      interface{}            `json:"data,omitempty"`
	OrderID   uint                   `json:"order_id,omitempty"`
	ItemID    uint                   `json:"item_id,omitempty"`
	TableID   uint                   `json:"table_id,omitempty"`
	UserID    uint                   `json:"user_id,omitempty"`
	Timestamp string                 `json:"timestamp"`
	Room      string                 `json:"room,omitempty"`
}

// ========================================
// NOTIFICATION CREATION
// ========================================

// CreateOrderNotification creates order notification
func (m *WebSocketManager) CreateOrderNotification(order Order) error {
	notification := Notification{
		Type:      "order",
		Action:    "created",
		Data:      order,
		OrderID:   order.ID,
		Timestamp: time.Now().Format(time.RFC3339),
		Room:      "all",
	}

	return m.Broadcast(notification)
}

// CreateOrderStatusNotification creates order status update notification
func (m *WebSocketManager) CreateOrderStatusNotification(orderID uint, oldStatus, newStatus string, order Order) error {
	notification := Notification{
		Type:      "order_status",
		Action:    "updated",
		OrderID:   orderID,
		Data:      map[string]interface{}{
			"old_status": oldStatus,
			"new_status": newStatus,
			"order":      order,
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Room:      "all",
	}

	// Send to specific rooms based on status
	if newStatus == "ready" {
		m.SendToRoom("kitchen", notification)
		m.SendToRoom("waiters", notification)
	} else if newStatus == "completed" {
		m.SendToRoom("pos", notification)
		m.SendToRoom("tables", notification)
	}

	return m.Broadcast(notification)
}

// CreateItemStatusNotification creates item status update notification
func (m *WebSocketManager) CreateItemStatusNotification(orderID, itemID uint, status string) error {
	notification := Notification{
		Type:      "item_status",
		Action:    "updated",
		OrderID:   orderID,
		ItemID:    itemID,
		Data: map[string]interface{}{
			"status": status,
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Room:      "kitchen", // Send to kitchen
	}

	return m.SendToRoom("kitchen", notification)
}

// CreateTableStatusNotification creates table status notification
func (m *WebSocketManager) CreateTableStatusNotification(tableID uint, oldStatus, newStatus string) error {
	notification := Notification{
		Type:      "table_status",
		Action:    "updated",
		TableID:   tableID,
		Data: map[string]interface{}{
			"old_status": oldStatus,
			"new_status": newStatus,
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Room:      "all",
	}

	return m.Broadcast(notification)
}

// CreatePaymentNotification creates payment notification
func (m *WebSocketManager) CreatePaymentNotification(orderID uint, payment Payment) error {
	notification := Notification{
		Type:      "payment",
		Action:    "created",
		OrderID:   orderID,
		Data:      payment,
		Timestamp: time.Now().Format(time.RFC3339),
		Room:      "all",
	}

	return m.Broadcast(notification)
}

// CreateDashboardUpdateNotification creates dashboard stats update
func (m *WebSocketManager) CreateDashboardUpdateNotification(stats DashboardStats) error {
	notification := Notification{
		Type:      "dashboard_update",
		Action:    "refresh",
		Data:      stats,
		Timestamp: time.Now().Format(time.RFC3339),
		Room:      "dashboard",
	}

	return m.SendToRoom("dashboard", notification)
}

// CreateLowStockNotification creates low stock alert
func (m *WebSocketManager) CreateLowStockNotification(item StockItem) error {
	notification := Notification{
		Type:      "low_stock",
		Action:    "alert",
		Data:      item,
		Timestamp: time.Now().Format(time.RFC3339),
		Room:      "inventory",
	}

	return m.Broadcast(notification)
}

// CreateShiftNotification creates shift notification
func (m *WebSocketManager) CreateShiftNotification(userID uint, action string) error {
	notification := Notification{
		Type:      "shift",
		Action:    action,
		UserID:    userID,
		Timestamp: time.Now().Format(time.RFC3339),
		Room:      "staff",
	}

	return m.SendToRoom("staff", notification)
}

// CreateKitchenTicketNotification sends new order to kitchen
func (m *WebSocketManager) CreateKitchenTicketNotification(order Order) error {
	notification := Notification{
		Type:      "kitchen_order",
		Action:    "created",
		Data:      order,
		Timestamp: time.Now().Format(time.RFC3339),
		Room:      "kitchen",
		Sound:     "new_order.mp3",
	}

	// Send to kitchen and play sound
	m.SendToRoom("kitchen", notification)

	// Send sound notification separately for alert
	soundNotification := map[string]interface{}{
		"type":      "sound_alert",
		"sound":     "new_order.mp3",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	return m.SendToRoom("kitchen", soundNotification)
}

// CreateCustomerNotification creates customer notification
func (m *WebSocketManager) CreateCustomerNotification(customer Customer) error {
	notification := Notification{
		Type:      "customer",
		Action:    "created",
		Data:      customer,
		Timestamp: time.Now().Format(time.RFC3339),
		Room:      "customers",
	}

	return m.Broadcast(notification)
}

// ========================================
// HELPER METHODS
// ========================================

// GetCurrentTime returns current time string
func GetCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// BroadcastMessage sends generic message to all clients
func (m *WebSocketManager) BroadcastMessage(messageType string, message string) error {
	notification := map[string]interface{}{
		"type":      messageType,
		"message":   message,
		"timestamp": GetCurrentTime(),
	}

	return m.Broadcast(notification)
}
