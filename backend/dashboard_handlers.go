package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ========================================
// DASHBOARD HANDLERS
// ========================================

// DashboardStats holds dashboard statistics
type DashboardStats struct {
	TotalRevenue      float64 `json:"total_revenue"`
	TotalOrders       int     `json:"total_orders"`
	TotalCustomers    int     `json:"total_customers"`
	AverageOrderValue float64 `json:"average_order_value"`
	DineInOrders      int     `json:"dine_in_orders"`
	TakeawayOrders    int     `json:"takeaway_orders"`
	DeliveryOrders     int     `json:"delivery_orders"`
	CashPayments      float64 `json:"cash_payments"`
	CardPayments      float64 `json:"card_payments"`
	WalletPayments    float64 `json:"wallet_payments"`
	LowStockItems     int     `json:"low_stock_items"`
	PendingOrders      int     `json:"pending_orders"`
	PreparingOrders   int     `json:"preparing_orders"`
	ReadyOrders        int     `json:"ready_orders"`
	ServedOrders       int     `json:"served_orders"`
	ActiveTables       int     `json:"active_tables"`
	AvailableTables   int     `json:"available_tables"`
	HourlyRevenue     []HourlyStat `json:"hourly_revenue"`
	TrendingItems      []TrendingItem `json:"trending_items"`
	SlowItems          []SlowItem `json:"slow_items"`
	TopStaff           []StaffStat `json:"top_staff"`
	BusyHours         []BusyHour `json:"busy_hours"`
}

// HourlyStat holds hourly revenue data
type HourlyStat struct {
	Hour   int     `json:"hour"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

// TrendingItem holds trending item data
type TrendingItem struct {
	ItemID       uint    `json:"item_id"`
	ItemName     string  `json:"item_name"`
	ItemNameAr   string  `json:"item_name_ar"`
	Category     string  `json:"category"`
	SoldCount    int     `json:"sold_count"`
	Revenue      float64 `json:"revenue"`
	Growth       float64 `json:"growth"`
}

// SlowItem holds slow selling item data
type SlowItem struct {
	ItemID      uint    `json:"item_id"`
	ItemName    string  `json:"item_name"`
	ItemNameAr  string  `json:"item_name_ar"`
	SoldCount   int     `json:"sold_count"`
	Revenue     float64 `json:"revenue"`
	LastSoldAt string  `json:"last_sold_at"`
}

// StaffStat holds staff performance data
type StaffStat struct {
	UserID      uint    `json:"user_id"`
	UserName    string  `json:"user_name"`
	OrdersCount int     `json:"orders_count"`
	TotalSales  float64 `json:"total_sales"`
	AverageOrder float64 `json:"average_order"`
	Rating      float64 `json:"rating"`
}

// BusyHour holds busy hour data
type BusyHour struct {
	Hour       int   `json:"hour"`
	Orders     int   `json:"orders"`
	Revenue    float64 `json:"revenue"`
	LoadFactor string `json:"load_factor"`
}

// ========================================
// DASHBOARD ENDPOINTS
// ========================================

// HandleGetDashboard returns complete dashboard stats
func (a *App) HandleGetDashboard(c *gin.Context) {
	// Get date range (default today)
	startDate := a.getStartDate(c)
	endDate := a.getEndDate(c)

	stats := DashboardStats{}

	// 1. Total Revenue & Orders
	a.DB.Model(&Order{}).
		Where("created_at >= ? AND created_at < ? AND payment_status = ?", startDate, endDate, "paid").
		Select("COALESCE(SUM(total), 0) as total_revenue").
		Select("COUNT(*) as total_orders").
		Scan(&stats.TotalRevenue, &stats.TotalOrders)

	// 2. Total Customers
	a.DB.Model(&Customer{}).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Count(&stats.TotalCustomers)

	// 3. Average Order Value
	if stats.TotalOrders > 0 {
		stats.AverageOrderValue = stats.TotalRevenue / float64(stats.TotalOrders)
	}

	// 4. Orders by Type
	a.DB.Model(&Order{}).
		Where("created_at >= ? AND created_at < ? AND payment_status = ?", startDate, endDate, "paid").
		Where("type = ?", "dine_in").Count(&stats.DineInOrders)
	a.DB.Model(&Order{}).
		Where("created_at >= ? AND created_at < ? AND payment_status = ?", startDate, endDate, "paid").
		Where("type = ?", "takeaway").Count(&stats.TakeawayOrders)
	a.DB.Model(&Order{}).
		Where("created_at >= ? AND created_at < ? AND payment_status = ?", startDate, endDate, "paid").
		Where("type = ?", "delivery").Count(&stats.DeliveryOrders)

	// 5. Payments by Method
	a.DB.Model(&Payment{}).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Where("method = ?", "cash").Select("COALESCE(SUM(amount), 0)").Scan(&stats.CashPayments)
	a.DB.Model(&Payment{}).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Where("method = ?", "card").Select("COALESCE(SUM(amount), 0)").Scan(&stats.CardPayments)
	a.DB.Model(&Payment{}).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Where("method = ?", "mobile_wallet").Select("COALESCE(SUM(amount), 0)").Scan(&stats.WalletPayments)

	// 6. Low Stock Items
	a.DB.Model(&StockItem{}).
		Where("is_low_stock = ? AND current_stock <= minimum_stock", true).
		Count(&stats.LowStockItems)

	// 7. Order Status Counts
	a.DB.Model(&Order{}).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Where("status = ?", "pending").Count(&stats.PendingOrders)
	a.DB.Model(&Order{}).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Where("status = ?", "preparing").Count(&stats.PreparingOrders)
	a.DB.Model(&Order{}).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Where("status = ?", "ready").Count(&stats.ReadyOrders)
	a.DB.Model(&Order{}).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Where("status = ?", "served").Count(&stats.ServedOrders)

	// 8. Table Status
	a.DB.Model(&Table{}).
		Where("status = ?", "occupied").Count(&stats.ActiveTables)
	a.DB.Model(&Table{}).
		Where("status = ?", "available").Count(&stats.AvailableTables)

	// 9. Hourly Revenue
	a.DB.Model(&Payment{}).
		Select("HOUR(created_at) as hour, COALESCE(SUM(amount), 0) as revenue").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group("HOUR(created_at)").
		Order("hour").
		Scan(&stats.HourlyRevenue)

	// 10. Trending Items (Top Selling)
	type ItemStat struct {
		ItemID      uint    `json:"item_id"`
		ItemName    string  `json:"item_name"`
		SoldCount   int     `json:"sold_count"`
		Revenue     float64 `json:"revenue"`
	}

	var itemStats []ItemStat
	a.DB.Model(&OrderItem{}).
		Select("menu_item_id as item_id, menu_item_name as item_name, SUM(quantity) as sold_count, SUM(quantity * unit_price) as revenue").
		Joins("LEFT JOIN orders ON orders.id = order_items.order_id").
		Where("orders.created_at >= ? AND orders.created_at < ?", startDate, endDate).
		Where("orders.payment_status = ?", "paid").
		Group("menu_item_id").
		Order("revenue DESC").
		Limit(10).
		Scan(&itemStats)

	for _, stat := range itemStats {
		var menuItem MenuItem
		a.DB.First(&menuItem, stat.ItemID)

		trending := TrendingItem{
			ItemID:     stat.ItemID,
			ItemName:   menuItem.Name,
			ItemNameAr: menuItem.NameAr,
			SoldCount:  stat.SoldCount,
			Revenue:    stat.Revenue,
			Growth:      stat.Revenue / float64(stat.SoldCount),
		}
		stats.TrendingItems = append(stats.TrendingItems, trending)
	}

	// 11. Slow Items (Low Selling)
	type SlowItemStat struct {
		ItemID      uint    `json:"item_id"`
		ItemName    string  `json:"item_name"`
		SoldCount   int     `json:"sold_count"`
		Revenue     float64 `json:"revenue"`
		LastSoldAt  string  `json:"last_sold_at"`
	}

	var slowItemStats []SlowItemStat
	a.DB.Model(&OrderItem{}).
		Select("menu_item_id as item_id, menu_item_name as item_name, COUNT(*) as sold_count, SUM(quantity * unit_price) as revenue, MAX(orders.created_at) as last_sold_at").
		Joins("LEFT JOIN orders ON orders.id = order_items.order_id").
		Where("orders.created_at >= ? AND orders.created_at < ?", startDate, endDate).
		Where("orders.payment_status = ?", "paid").
		Group("menu_item_id").
		Having("COUNT(*) > 0").
		Order("revenue ASC").
		Limit(10).
		Scan(&slowItemStats)

	for _, stat := range slowItemStats {
		var menuItem MenuItem
		a.DB.First(&menuItem, stat.ItemID)

		slow := SlowItem{
			ItemID:     stat.ItemID,
			ItemName:   menuItem.Name,
			ItemNameAr: menuItem.NameAr,
			SoldCount:  stat.SoldCount,
			Revenue:    stat.Revenue,
			LastSoldAt: stat.LastSoldAt,
		}
		stats.SlowItems = append(stats.SlowItems, slow)
	}

	// 12. Top Staff Performance
	a.DB.Model(&Order{}).
		Select("user_id, user_id, COUNT(*) as orders_count, COALESCE(SUM(total), 0) as total_sales").
		Where("created_at >= ? AND created_at < ? AND payment_status = ?", startDate, endDate, "paid").
		Group("user_id").
		Order("total_sales DESC").
		Limit(5).
		Scan(&stats.TopStaff)

	// 13. Busy Hours
	type BusyHourStat struct {
		Hour   int   `json:"hour"`
		Orders int   `json:"orders"`
		Revenue float64 `json:"revenue"`
	}

	var busyHourStats []BusyHourStat
	a.DB.Model(&Order{}).
		Select("HOUR(created_at) as hour, COUNT(*) as orders, COALESCE(SUM(total), 0) as revenue").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group("HOUR(created_at)").
		Order("orders DESC").
		Scan(&busyHourStats)

	maxOrders := 0
	if len(busyHourStats) > 0 {
		maxOrders = busyHourStats[0].Orders
	}

	for _, stat := range busyHourStats {
		loadFactor := "Low"
		if stat.Orders > maxOrders*0.75 {
			loadFactor = "High"
		} else if stat.Orders > maxOrders*0.5 {
			loadFactor = "Medium"
		}

		stats.BusyHours = append(stats.BusyHours, BusyHour{
			Hour:       stat.Hour,
			Orders:     stat.Orders,
			Revenue:    stat.Revenue,
			LoadFactor: loadFactor,
		})
	}

	c.JSON(http.StatusOK, stats)
}

// HandleGetDashboardStats returns dashboard stats (separate endpoint for charts)
func (a *App) HandleGetDashboardStats(c *gin.Context) {
	dateFilter := c.Query("date_filter") // "today", "week", "month"

	startDate, endDate := a.getDateRange(dateFilter)

	stats := map[string]interface{}{}

	// Sales over time (line chart)
	type SalesData struct {
		Date   string `json:"date"`
		Revenue float64 `json:"revenue"`
		Orders  int     `json:"orders"`
	}

	var salesData []SalesData
	a.DB.Model(&Order{}).
		Select("DATE(created_at) as date, COALESCE(SUM(total), 0) as revenue, COUNT(*) as orders").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Where("payment_status = ?", "paid").
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&salesData)

	stats["sales_over_time"] = salesData

	// Revenue by category (pie chart)
	type CategoryRevenue struct {
		Category   string  `json:"category"`
		Revenue    float64 `json:"revenue"`
		Orders     int     `json:"orders"`
	}

	var categoryRevenue []CategoryRevenue
	a.DB.Model(&Order{}).
		Select("categories.name as category, COALESCE(SUM(orders.total), 0) as revenue, COUNT(*) as orders").
		Joins("LEFT JOIN menu_items ON menu_items.id = order_items.menu_item_id").
		Joins("LEFT JOIN orders ON orders.id = order_items.order_id").
		Joins("LEFT JOIN categories ON categories.id = menu_items.category_id").
		Where("orders.created_at >= ? AND orders.created_at < ? AND orders.payment_status = ?", startDate, endDate, "paid").
		Group("categories.id").
		Order("revenue DESC").
		Scan(&categoryRevenue)

	stats["revenue_by_category"] = categoryRevenue

	// Revenue by payment method (doughnut chart)
	type PaymentRevenue struct {
		Method  string  `json:"method"`
		Revenue float64 `json:"revenue"`
	}

	var paymentRevenue []PaymentRevenue
	a.DB.Model(&Payment{}).
		Select("method, COALESCE(SUM(amount), 0) as revenue").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group("method").
		Scan(&paymentRevenue)

	stats["revenue_by_payment"] = paymentRevenue

	// Top items (horizontal bar chart)
	stats["top_items"] = []TrendingItem{}

	type TopItem struct {
		ItemID       uint    `json:"item_id"`
		ItemName     string  `json:"item_name"`
		ItemNameAr   string  `json:"item_name_ar"`
		SoldCount    int     `json:"sold_count"`
		Revenue      float64 `json:"revenue"`
	}

	var topItems []TopItem
	a.DB.Model(&OrderItem{}).
		Select("menu_item_id as item_id, menu_item_name as item_name, SUM(quantity) as sold_count, SUM(quantity * unit_price) as revenue").
		Joins("LEFT JOIN orders ON orders.id = order_items.order_id").
		Where("orders.created_at >= ? AND orders.created_at < ? AND orders.payment_status = ?", startDate, endDate, "paid").
		Group("menu_item_id").
		Order("revenue DESC").
		Limit(10).
		Scan(&topItems)

	for _, item := range topItems {
		var menuItem MenuItem
		a.DB.First(&menuItem, item.ItemID)

		trending := TrendingItem{
			ItemID:     item.ItemID,
			ItemName:   menuItem.Name,
			ItemNameAr: menuItem.NameAr,
			SoldCount:  item.SoldCount,
			Revenue:    item.Revenue,
			Growth:      item.Revenue / float64(item.SoldCount),
		}
		stats["top_items"] = append(stats["top_items"].([]TrendingItem), trending)
	}

	// Kitchen load
	type KitchenLoad struct {
		Hour       int   `json:"hour"`
		Pending    int   `json:"pending"`
		Preparing  int   `json:"preparing"`
		Ready      int   `json:"ready"`
	}

	var kitchenLoad []KitchenLoad
	a.DB.Model(&OrderItem{}).
		Select("HOUR(orders.created_at) as hour").
		Joins("LEFT JOIN orders ON orders.id = order_items.order_id").
		Where("orders.created_at >= ? AND orders.created_at < ?", startDate, endDate).
		Group("HOUR(orders.created_at)").
		Order("hour").
		Scan(&kitchenLoad)

	for _, hour := range kitchenLoad {
		var hourStat KitchenLoad
		hourStat.Hour = hour.Hour

		a.DB.Model(&OrderItem{}).
			Joins("LEFT JOIN orders ON orders.id = order_items.order_id").
			Where("HOUR(orders.created_at) = ? AND orders.status = ?", hour.Hour, "pending").
			Count(&hourStat.Pending)

		a.DB.Model(&OrderItem{}).
			Joins("LEFT JOIN orders ON orders.id = order_items.order_id").
			Where("HOUR(orders.created_at) = ? AND orders.status = ?", hour.Hour, "preparing").
			Count(&hourStat.Preparing)

		a.DB.Model(&OrderItem{}).
			Joins("LEFT JOIN orders ON orders.id = order_items.order_id").
			Where("HOUR(orders.created_at) = ? AND orders.status = ?", hour.Hour, "ready").
			Count(&hourStat.Ready)

		kitchenLoad = append(kitchenLoad, hourStat)
	}

	stats["kitchen_load"] = kitchenLoad

	// Wait time metrics
	type WaitTime struct {
		AvgWaitTime float64 `json:"avg_wait_time"`
		MaxWaitTime float64 `json:"max_wait_time"`
	}

	var waitTime WaitTime
	a.DB.Model(&Order{}).
		Select("AVG(TIMESTAMPDIFF(MINUTE, created_at, completed_at)) as avg_wait_time, MAX(TIMESTAMPDIFF(MINUTE, created_at, completed_at)) as max_wait_time").
		Where("created_at >= ? AND completed_at IS NOT NULL", startDate).
		Scan(&waitTime)

	stats["wait_time_metrics"] = waitTime

	c.JSON(http.StatusOK, stats)
}

// HandleGetRecentOrders returns recent orders for dashboard
func (a *App) HandleGetRecentOrders(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	page := c.DefaultQuery("page", "1")

	var orders []Order

	query := a.DB.Preload("Table").Preload("Items").Order("created_at DESC")

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	offset := (getInt(page) - 1) * getInt(limit)
	query = query.Offset(offset).Limit(getInt(limit))

	if err := query.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// HandleGetLowStockDashboard returns low stock items for dashboard
func (a *App) HandleGetLowStockDashboard(c *gin.Context) {
	var items []StockItem

	if err := a.DB.Where("is_low_stock = ?", true).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch low stock items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

// ========================================
// HELPER FUNCTIONS
// ========================================

// getStartDate extracts start date from query params
func (a *App) getStartDate(c *gin.Context) time.Time {
	dateFilter := c.Query("date_filter")
	today := time.Now().Truncate(24 * time.Hour)

	switch dateFilter {
	case "today":
		return today
	case "week":
		return today.AddDate(-7 * 24 * time.Hour)
	case "month":
		return today.AddDate(-30 * 24 * time.Hour)
	case "year":
		return today.AddDate(-365 * 24 * time.Hour)
	default:
		return today
	}
}

// getEndDate extracts end date from query params
func (a *App) getEndDate(c *gin.Context) time.Time {
	dateFilter := c.Query("date_filter")
	today := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour) // Tomorrow

	switch dateFilter {
	case "today":
		return today
	case "week":
		return today
	case "month":
		return today
	case "year":
		return today
	default:
		return today
	}
}

// getDateRange returns date range based on filter
func (a *App) getDateRange(dateFilter string) (time.Time, time.Time) {
	today := time.Now().Truncate(24 * time.Hour)

	switch dateFilter {
	case "today":
		return today, today.Add(24 * time.Hour)
	case "week":
		return today.AddDate(-7 * 24 * time.Hour), today
	case "month":
		return today.AddDate(-30 * 24 * time.Hour), today
	default:
		return today, today.Add(24 * time.Hour)
	}
}
