package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

// ========================================
// NOTIFICATION SERVICE
// ========================================

// NotificationService handles sending notifications
type NotificationService struct {
	WebSocket *WebSocketManager
	Database *gorm.DB
}

// NewNotificationService creates new notification service
func NewNotificationService(ws *WebSocketManager, db *gorm.DB) *NotificationService {
	return &NotificationService{
		WebSocket: ws,
		Database:  db,
	}
}

// SendOrderNotification sends order notification to relevant clients
func (n *NotificationService) SendOrderNotification(orderID uint, order Order) error {
	notification := map[string]interface{}{
		"type":     "order",
		"action":   "created",
		"order_id": order.ID,
		"data":     order,
	}

	// Broadcast to all connected clients
	n.WebSocket.Broadcast(notification)
	return nil
}

// SendOrderStatusUpdate sends order status update
func (n *NotificationService) SendOrderStatusUpdate(orderID uint, oldStatus, newStatus string) error {
	notification := map[string]interface{}{
		"type":       "order_status",
		"action":     "updated",
		"order_id":   orderID,
		"old_status": oldStatus,
		"new_status": newStatus,
		"timestamp":  getCurrentTime(),
	}

	n.WebSocket.Broadcast(notification)

	// Send to specific room if status is "ready"
	if newStatus == "ready" {
		n.WebSocket.SendToRoom("kitchen", notification)
	}

	// Send to POS if order is completed
	if newStatus == "completed" || newStatus == "cancelled" {
		n.WebSocket.SendToRoom("pos", notification)
	}

	return nil
}

// SendItemStatusUpdate sends item status update
func (n *NotificationService) SendItemStatusUpdate(orderID, itemID uint, status string) error {
	notification := map[string]interface{}{
		"type":      "item_status",
		"action":    "updated",
		"order_id":  orderID,
		"item_id":   itemID,
		"status":    status,
		"timestamp": getCurrentTime(),
	}

	// Send to kitchen
	n.WebSocket.SendToRoom("kitchen", notification)

	return nil
}

// SendTableStatusUpdate sends table status update
func (n *NotificationService) SendTableStatusUpdate(tableID uint, oldStatus, newStatus string) error {
	notification := map[string]interface{}{
		"type":       "table_status",
		"action":     "updated",
		"table_id":   tableID,
		"old_status": oldStatus,
		"new_status": newStatus,
		"timestamp":  getCurrentTime(),
	}

	// Broadcast to POS and Tables
	n.WebSocket.SendToRoom("pos", notification)
	n.WebSocket.SendToRoom("tables", notification)

	return nil
}

// SendPaymentNotification sends payment notification
func (n *NotificationService) SendPaymentNotification(orderID uint, payment Payment) error {
	notification := map[string]interface{}{
		"type":      "payment",
		"action":    "created",
		"order_id":  orderID,
		"data":      payment,
		"timestamp": getCurrentTime(),
	}

	// Broadcast to POS
	n.WebSocket.SendToRoom("pos", notification)

	// Send to manager if large amount
	if payment.Amount > 1000 {
		n.WebSocket.SendToRoom("managers", notification)
	}

	return nil
}

// SendLowStockAlert sends low stock alert
func (n *NotificationService) SendLowStockAlert(stockItem StockItem) error {
	notification := map[string]interface{}{
		"type":       "low_stock",
		"action":     "alert",
		"data":       stockItem,
		"timestamp":  getCurrentTime(),
	}

	// Broadcast to all
	n.WebSocket.Broadcast(notification)

	return nil
}

// SendNewCustomerNotification sends new customer notification
func (n *NotificationService) SendNewCustomerNotification(customer Customer) error {
	notification := map[string]interface{}{
		"type":       "customer",
		"action":     "created",
		"data":       customer,
		"timestamp":  getCurrentTime(),
	}

	n.WebSocket.SendToRoom("managers", notification)
	return nil
}

// SendReservationNotification sends reservation notification
func (n *NotificationService) SendReservationNotification(reservation Reservation) error {
	notification := map[string]interface{}{
		"type":       "reservation",
		"action":     "created",
		"data":       reservation,
		"timestamp":  getCurrentTime(),
	}

	// Send to waiters
	n.WebSocket.SendToRoom("waiters", notification)
	n.WebSocket.SendToRoom("tables", notification)

	return nil
}

// SendShiftNotification sends shift notification
func (n *NotificationService) SendShiftNotification(userID uint, action string) error {
	notification := map[string]interface{}{
		"type":       "shift",
		"action":     action,
		"user_id":    userID,
		"timestamp":  getCurrentTime(),
	}

	n.WebSocket.SendToRoom("managers", notification)
	return nil
}

// SendKitchenTicketNotification sends new order to kitchen
func (n *NotificationService) SendKitchenTicketNotification(order Order) error {
	notification := map[string]interface{}{
		"type":       "kitchen_order",
		"action":     "created",
		"order":      order,
		"timestamp":  getCurrentTime(),
	}

	// Send to kitchen
	n.WebSocket.SendToRoom("kitchen", notification)

	// Play sound alert
	n.WebSocket.SendToRoom("kitchen", map[string]interface{}{
		"type": "sound_alert",
		"sound": "new_order.mp3",
	})

	return nil
}

// BroadcastDashboardUpdate sends dashboard stat updates
func (n *NotificationService) BroadcastDashboardUpdate(stats DashboardStats) error {
	notification := map[string]interface{}{
		"type":       "dashboard_update",
		"action":     "refresh",
		"data":       stats,
		"timestamp": getCurrentTime(),
	}

	n.WebSocket.Broadcast(notification)
	return nil
}

func getCurrentTime() string {
	return getCurrentTime()
}
