package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ========================================
// AUTHENTICATION HANDLERS
// ========================================

// HandleLogin handles user login
func (a *App) HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	// Find user by email
	var user User
	if err := a.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Account is deactivated"})
		return
	}

	// Generate token
	token, err := a.GenerateToken(fmt.Sprintf("%d", user.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	// Update last login
	a.DB.Model(&user).Update("LastLogin", a.GetCurrentTime())

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

// HandleLogout handles user logout
func (a *App) HandleLogout(c *gin.Context) {
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

// HandleRefreshToken handles token refresh
func (a *App) HandleRefreshToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Missing token"})
		return
	}

	// Validate and refresh token
	// In production, use proper JWT refresh logic
	userID, err := a.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token"})
		return
	}

	// Generate new token
	newToken, err := a.GenerateToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}

// HandleGetCurrentUser returns current logged in user
func (a *App) HandleGetCurrentUser(c *gin.Context) {
	userID := a.GetUserIDFromContext(c)

	var user User
	if err := a.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ========================================
// SETTINGS HANDLERS
// ========================================

// HandleGetSettings returns restaurant settings
func (a *App) HandleGetSettings(c *gin.Context) {
	var settings RestaurantSettings
	if err := a.DB.First(&settings).Error; err != nil {
		// Return default settings
		settings = RestaurantSettings{
			Name:           "مطعم",
			NameAr:         "مطعم",
			Currency:        "EGP",
			CurrencySymbol:  "ج.م",
			TaxRate:         0.14,
			ServiceCharge:    0.10,
			Language:        "ar",
			ThemeColor:      "#10b981",
			IsOpen:          true,
		}
	}

	c.JSON(http.StatusOK, settings)
}

// HandleUpdateSettings updates restaurant settings
func (a *App) HandleUpdateSettings(c *gin.Context) {
	var settings RestaurantSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	if err := a.DB.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Settings updated successfully",
		Data:    settings,
	})
}

// ========================================
// MENU HANDLERS - CATEGORIES
// ========================================

// HandleGetCategories returns all categories
func (a *App) HandleGetCategories(c *gin.Context) {
	var categories []Category

	query := a.DB.Order("display_order ASC")
	if err := query.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// HandleCreateCategory creates a new category
func (a *App) HandleCreateCategory(c *gin.Context) {
	var category Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	if err := a.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Message: "Category created successfully",
		Data:    category,
	})
}

// HandleGetCategory returns a single category
func (a *App) HandleGetCategory(c *gin.Context) {
	id := c.Param("id")

	var category Category
	if err := a.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// HandleUpdateCategory updates a category
func (a *App) HandleUpdateCategory(c *gin.Context) {
	id := c.Param("id")

	var category Category
	if err := a.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Category not found"})
		return
	}

	var updates Category
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	if err := a.DB.Model(&category).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Category updated successfully",
		Data:    category,
	})
}

// HandleDeleteCategory deletes a category
func (a *App) HandleDeleteCategory(c *gin.Context) {
	id := c.Param("id")

	if err := a.DB.Delete(&Category{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Category deleted successfully",
	})
}

// ========================================
// MENU HANDLERS - MENU ITEMS
// ========================================

// HandleGetMenuItems returns all menu items
func (a *App) HandleGetMenuItems(c *gin.Context) {
	var items []MenuItem

	query := a.DB.Preload("Category")
	if err := query.Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch menu items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

// HandleCreateMenuItem creates a new menu item
func (a *App) HandleCreateMenuItem(c *gin.Context) {
	var item MenuItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	if err := a.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create menu item"})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Message: "Menu item created successfully",
		Data:    item,
	})
}

// HandleGetMenuItem returns a single menu item
func (a *App) HandleGetMenuItem(c *gin.Context) {
	id := c.Param("id")

	var item MenuItem
	if err := a.DB.Preload("Category").First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Menu item not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// HandleUpdateMenuItem updates a menu item
func (a *App) HandleUpdateMenuItem(c *gin.Context) {
	id := c.Param("id")

	var item MenuItem
	if err := a.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Menu item not found"})
		return
	}

	var updates MenuItem
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	if err := a.DB.Model(&item).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update menu item"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Menu item updated successfully",
		Data:    item,
	})
}

// HandleDeleteMenuItem deletes a menu item
func (a *App) HandleDeleteMenuItem(c *gin.Context) {
	id := c.Param("id")

	if err := a.DB.Delete(&MenuItem{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete menu item"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Menu item deleted successfully",
	})
}

// ========================================
// ORDERS HANDLERS
// ========================================

// HandleGetOrders returns all orders
func (a *App) HandleGetOrders(c *gin.Context) {
	var orders []Order

	query := a.DB.Preload("Table").Preload("Items").Order("created_at DESC")
	
	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by date range
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	// Pagination
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "50")
	query = query.Offset((getInt(page) - 1) * getInt(limit)).Limit(getInt(limit))

	if err := query.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// HandleCreateOrder creates a new order
func (a *App) HandleCreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	// Get user ID
	userID := a.GetUserIDFromContext(c)
	var user User
	a.DB.First(&user, userID)

	// Generate order number
	orderNumber := fmt.Sprintf("ORD-%d", a.GetCurrentTimestamp())

	// Create order
	order := Order{
		OrderNumber:    orderNumber,
		TableID:        req.TableID,
		UserID:         user.ID,
		Type:           req.Type,
		Priority:       req.Priority,
		CustomerName:   req.CustomerName,
		CustomerPhone:  req.CustomerPhone,
		CustomerAddress: req.CustomerAddress,
		Notes:          req.Notes,
		KitchenNotes:   req.KitchenNotes,
		Status:         "pending",
		PaymentStatus:  "unpaid",
	}

	// Calculate totals
	var subtotal, total float64
	for _, itemReq := range req.Items {
		var menuItem MenuItem
		a.DB.First(&menuItem, itemReq.MenuItemID)
		subtotal += float64(itemReq.Quantity) * menuItem.Price
	}

	// Get settings
	var settings RestaurantSettings
	a.DB.First(&settings)

	taxAmount := subtotal * settings.TaxRate
	serviceCharge := subtotal * settings.ServiceCharge
	total = subtotal + taxAmount + serviceCharge

	order.Subtotal = subtotal
	order.TaxAmount = taxAmount
	order.ServiceCharge = serviceCharge
	order.Total = total

	// Create order in DB
	if err := a.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create order"})
		return
	}

	// Create order items
	for _, itemReq := range req.Items {
		var menuItem MenuItem
		a.DB.First(&menuItem, itemReq.MenuItemID)

		orderItem := OrderItem{
			OrderID:      order.ID,
			MenuItemID:   itemReq.MenuItemID,
			MenuItemName: menuItem.Name,
			Quantity:     itemReq.Quantity,
			UnitPrice:    menuItem.Price,
			Modifiers:    itemReq.Modifiers,
			Notes:        itemReq.Notes,
			Status:       "pending",
		}

		a.DB.Create(&orderItem)

		// Update item order count
		a.DB.Model(&menuItem).UpdateColumn("order_count", gorm.Expr("order_count + 1"))
	}

	// Update table status if table is assigned
	if req.TableID != nil {
		a.DB.Model(&Table{}).Where("id = ?", *req.TableID).Updates(map[string]interface{}{
			"status":           "occupied",
			"current_order_id": order.ID,
		})
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Message: "Order created successfully",
		Data:    order,
	})
}

// HandleGetOrder returns a single order
func (a *App) HandleGetOrder(c *gin.Context) {
	id := c.Param("id")

	var order Order
	if err := a.DB.Preload("Table").Preload("Items").Preload("Payments").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// HandleUpdateOrder updates an order
func (a *App) HandleUpdateOrder(c *gin.Context) {
	id := c.Param("id")

	var order Order
	if err := a.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		return
	}

	var updates Order
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request", Message: err.Error()})
		return
	}

	if err := a.DB.Model(&order).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update order"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Order updated successfully",
		Data:    order,
	})
}

// HandleDeleteOrder deletes an order
func (a *App) HandleDeleteOrder(c *gin.Context) {
	id := c.Param("id")

	if err := a.DB.Delete(&Order{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete order"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Order deleted successfully",
	})
}

// HandleUpdateOrderStatus updates order status
func (a *App) HandleUpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	var order Order
	if err := a.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		return
	}

	updates := map[string]interface{}{"status": req.Status}

	// Set timestamps based on status
	switch req.Status {
	case "confirmed":
		updates["started_at"] = a.GetCurrentTime()
	case "completed":
		updates["completed_at"] = a.GetCurrentTime()
		updates["payment_status"] = "paid"
	case "cancelled":
		updates["cancelled_at"] = a.GetCurrentTime()
	}

	if err := a.DB.Model(&order).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update order status"})
		return
	}

	// Free table if order is completed or cancelled
	if req.Status == "completed" || req.Status == "cancelled" {
		if order.TableID != nil {
			a.DB.Model(&Table{}).Where("id = ?", *order.TableID).Updates(map[string]interface{}{
				"status":           "available",
				"current_order_id": nil,
			})
		}
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Order status updated successfully",
	})
}

// HandleAddOrderItem adds item to existing order
func (a *App) HandleAddOrderItem(c *gin.Context) {
	id := c.Param("id")

	var req CreateOrderItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	var order Order
	if err := a.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		return
	}

	var menuItem MenuItem
	a.DB.First(&menuItem, req.MenuItemID)

	orderItem := OrderItem{
		OrderID:      order.ID,
		MenuItemID:   req.MenuItemID,
		MenuItemName: menuItem.Name,
		Quantity:     req.Quantity,
		UnitPrice:    menuItem.Price,
		Modifiers:    req.Modifiers,
		Notes:        req.Notes,
		Status:       "pending",
	}

	if err := a.DB.Create(&orderItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to add item to order"})
		return
	}

	// Recalculate order totals
	a.recalculateOrderTotal(order.ID)

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Message: "Item added to order",
		Data:    orderItem,
	})
}

// HandleUpdateOrderItem updates an order item
func (a *App) HandleUpdateOrderItem(c *gin.Context) {
	id := c.Param("id")
	itemID := c.Param("itemId")

	var updates OrderItem
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	var orderItem OrderItem
	if err := a.DB.Where("order_id = ? AND id = ?", id, itemID).First(&orderItem).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order item not found"})
		return
	}

	if err := a.DB.Model(&orderItem).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update order item"})
		return
	}

	// Recalculate order totals
	a.recalculateOrderTotal(id)

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Order item updated successfully",
	})
}

// HandleDeleteOrderItem removes item from order
func (a *App) HandleDeleteOrderItem(c *gin.Context) {
	id := c.Param("id")
	itemID := c.Param("itemId")

	if err := a.DB.Where("order_id = ? AND id = ?", id, itemID).Delete(&OrderItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete order item"})
		return
	}

	// Recalculate order totals
	a.recalculateOrderTotal(id)

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Order item deleted successfully",
	})
}

// ========================================
// TABLES HANDLERS
// ========================================

// HandleGetTables returns all tables
func (a *App) HandleGetTables(c *gin.Context) {
	var tables []Table

	if err := a.DB.Order("number ASC").Find(&tables).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch tables"})
		return
	}

	c.JSON(http.StatusOK, tables)
}

// HandleCreateTable creates a new table
func (a *App) HandleCreateTable(c *gin.Context) {
	var table Table
	if err := c.ShouldBindJSON(&table); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	if err := a.DB.Create(&table).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create table"})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Message: "Table created successfully",
		Data:    table,
	})
}

// HandleGetTable returns a single table
func (a *App) HandleGetTable(c *gin.Context) {
	id := c.Param("id")

	var table Table
	if err := a.DB.First(&table, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Table not found"})
		return
	}

	c.JSON(http.StatusOK, table)
}

// HandleUpdateTable updates a table
func (a *App) HandleUpdateTable(c *gin.Context) {
	id := c.Param("id")

	var table Table
	if err := a.DB.First(&table, id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Table not found"})
		return
	}

	var updates Table
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	if err := a.DB.Model(&table).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update table"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Table updated successfully",
		Data:    table,
	})
}

// HandleDeleteTable deletes a table
func (a *App) HandleDeleteTable(c *gin.Context) {
	id := c.Param("id")

	if err := a.DB.Delete(&Table{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete table"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Table deleted successfully",
	})
}

// HandleUpdateTableStatus updates table status
func (a *App) HandleUpdateTableStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	if err := a.DB.Model(&Table{}).Where("id = ?", id).Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update table status"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Table status updated successfully",
	})
}

// HandleTransferTable transfers order from one table to another
func (a *App) HandleTransferTable(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		ToTableID uint `json:"to_table_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	// Free old table
	a.DB.Model(&Table{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":           "available",
		"current_order_id": nil,
	})

	// Assign to new table
	a.DB.Model(&Table{}).Where("id = ?", req.ToTableID).Updates(map[string]interface{}{
		"status": "occupied",
	})

	// Update order
	a.DB.Model(&Order{}).Where("table_id = ?", id).Update("table_id", req.ToTableID)

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Order transferred successfully",
	})
}

// ========================================
// HELPER FUNCTIONS
// ========================================

func (a *App) recalculateOrderTotal(orderID uint) {
	// Get order items
	var items []OrderItem
	a.DB.Where("order_id = ?", orderID).Find(&items)

	// Calculate subtotal
	var subtotal float64
	for _, item := range items {
		subtotal += float64(item.Quantity) * item.UnitPrice
	}

	// Get settings for tax rate
	var settings RestaurantSettings
	a.DB.First(&settings)

	taxAmount := subtotal * settings.TaxRate
	serviceCharge := subtotal * settings.ServiceCharge
	total := subtotal + taxAmount + serviceCharge

	// Update order
	a.DB.Model(&Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"subtotal":       subtotal,
		"tax_amount":     taxAmount,
		"service_charge": serviceCharge,
		"total":          total,
	})
}

func (a *App) GetCurrentTimestamp() int64 {
	return a.DB.NowFunc()().Unix()
}

func getInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

// Placeholders for handlers not yet implemented
func (a *App) HandleCreateModifier(c *gin.Context)         {}
func (a *App) HandleUpdateModifier(c *gin.Context)         {}
func (a *App) HandleDeleteModifier(c *gin.Context)         {}
func (a *App) HandleGetModifiers(c *gin.Context)           {}
func (a *App) HandleGetCombos(c *gin.Context)             {}
func (a *App) HandleCreateCombo(c *gin.Context)           {}
func (a *App) HandleUpdateCombo(c *gin.Context)           {}
func (a *App) HandleDeleteCombo(c *gin.Context)           {}
func (a *App) HandleGetPayments(c *gin.Context)           {}
func (a *App) HandleCreatePayment(c *gin.Context)         {}
func (a *App) HandleRefundPayment(c *gin.Context)          {}
func (a *App) HandleDailyReport(c *gin.Context)            {}
func (a *App) HandleWeeklyReport(c *gin.Context)           {}
func (a *App) HandleMonthlyReport(c *gin.Context)          {}
func (a *App) HandleItemsReport(c *gin.Context)            {}
func (a *App) HandleCategoriesReport(c *gin.Context)        {}
func (a *App) HandlePaymentsReport(c *gin.Context)         {}
func (a *App) HandleStaffReport(c *gin.Context)            {}
func (a *App) HandleExportReport(c *gin.Context)            {}
func (a *App) HandleGetStockItems(c *gin.Context)          {}
func (a *App) HandleCreateStockItem(c *gin.Context)        {}
func (a *App) HandleUpdateStockItem(c *gin.Context)        {}
func (a *App) HandleDeleteStockItem(c *gin.Context)        {}
func (a *App) HandleGetStockMovements(c *gin.Context)       {}
func (a *App) HandleAddStockMovement(c *gin.Context)        {}
func (a *App) HandleGetLowStockAlerts(c *gin.Context)     {}
func (a *App) HandleGetStaff(c *gin.Context)               {}
func (a *App) HandleCreateStaff(c *gin.Context)            {}
func (a *App) HandleUpdateStaff(c *gin.Context)            {}
func (a *App) HandleDeleteStaff(c *gin.Context)            {}
func (a *App) HandleStartShift(c *gin.Context)             {}
func (a *App) HandleEndShift(c *gin.Context)               {}
func (a *App) HandleGetShifts(c *gin.Context)               {}
func (a *App) HandleGetCustomers(c *gin.Context)            {}
func (a *App) HandleCreateCustomer(c *gin.Context)         {}
func (a *App) HandleGetCustomer(c *gin.Context)             {}
func (a *App) HandleUpdateCustomer(c *gin.Context)          {}
func (a *App) HandleDeleteCustomer(c *gin.Context)          {}
func (a *App) HandleAddLoyaltyPoints(c *gin.Context)       {}
func (a *App) HandleGetCustomerHistory(c *gin.Context)      {}
func (a *App) HandleGetReservations(c *gin.Context)         {}
func (a *App) HandleCreateReservation(c *gin.Context)        {}
func (a *App) HandleGetReservation(c *gin.Context)            {}
func (a *App) HandleUpdateReservation(c *gin.Context)         {}
func (a *App) HandleDeleteReservation(c *gin.Context)         {}
func (a *App) HandleUpdateReservationStatus(c *gin.Context)  {}
func (a *App) HandleGetDiscounts(c *gin.Context)           {}
func (a *App) HandleCreateDiscount(c *gin.Context)          {}
func (a *App) HandleUpdateDiscount(c *gin.Context)          {}
func (a *App) HandleDeleteDiscount(c *gin.Context)          {}
func (a *App) HandleActivateDiscount(c *gin.Context)         {}
func (a *App) HandleDeactivateDiscount(c *gin.Context)       {}
func (a *App) HandleGetPrinters(c *gin.Context)            {}
func (a *App) HandleCreatePrinter(c *gin.Context)            {}
func (a *App) HandleUpdatePrinter(c *gin.Context)            {}
func (a *App) HandleDeletePrinter(c *gin.Context)            {}
func (a *App) HandleTestPrinter(c *gin.Context)             {}
func (a *App) HandleGetReceiptTemplates(c *gin.Context)     {}
func (a *App) HandleCreateReceiptTemplate(c *gin.Context)    {}
func (a *App) HandleUpdateReceiptTemplate(c *gin.Context)    {}
func (a *App) HandleDeleteReceiptTemplate(c *gin.Context)    {}
func (a *App) HandleSetDefaultReceiptTemplate(c *gin.Context) {}
func (a *App) HandlePrintReceipt(c *gin.Context)             {}
func (a *App) HandlePrintKitchen(c *gin.Context)            {}
func (a *App) HandlePrintBar(c *gin.Context)                {}
func (a *App) HandleSendWhatsAppReceipt(c *gin.Context)       {}
func (a *App) HandleSendWhatsAppDailyReport(c *gin.Context) {}
func (a *App) HandleTestWhatsApp(c *gin.Context)             {}
func (a *App) HandleGetDashboard(c *gin.Context)             {}
func (a *App) HandleGetDashboardStats(c *gin.Context)        {}
func (a *App) HandleGetRecentOrders(c *gin.Context)          {}
func (a *App) HandleGetLowStockDashboard(c *gin.Context)     {}
