package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// App structure
type App struct {
	DB                  *gorm.DB
	Server              *gin.Engine
	Config              *Config
	WSManager           *WebSocketManager
	NotificationService *NotificationService
	Runtime             *wails.Runtime
}

// Config structure
type Config struct {
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
	}
	Server struct {
		Port int
		Host string
	}
	JWT struct {
		Secret     string
		Expiration time.Duration
	}
	WhatsApp struct {
		APIURL  string
		APIKey  string
		Enabled bool
	}
	Email struct {
		SMTPHost     string
		SMTPPort     int
		SMTPUser     string
		SMTPPassword string
		SMTPFrom     string
		Enabled       bool
	}
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	_ = godotenv.Load()

	config := &Config{
		Database: struct {
			Host     string
			Port     int
			User     string
			Password string
			Name     string
		}{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 3306),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "restaurant_pos"),
		},
		Server: struct {
			Port int
			Host string
		}{
			Port: getEnvInt("SERVER_PORT", 8080),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		JWT: struct {
			Secret     string
			Expiration time.Duration
		}{
			Secret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			Expiration: time.Hour * 24,
		},
		WhatsApp: struct {
			APIURL  string
			APIKey  string
			Enabled bool
		}{
			APIURL:  getEnv("WHATSAPP_API_URL", ""),
			APIKey:  getEnv("WHATSAPP_API_KEY", ""),
			Enabled: getEnv("WHATSAPP_ENABLED", "true") == "true",
		},
		Email: struct {
			SMTPHost     string
			SMTPPort     int
			SMTPUser     string
			SMTPPassword string
			SMTPFrom     string
			Enabled       bool
		}{
			SMTPHost:     getEnv("SMTP_HOST", ""),
			SMTPPort:     getEnvInt("SMTP_PORT", 587),
			SMTPUser:     getEnv("SMTP_USER", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			SMTPFrom:     getEnv("SMTP_FROM", ""),
			Enabled:       getEnv("EMAIL_ENABLED", "false") == "true",
		},
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intVal int
		fmt.Sscanf(value, "%d", &intVal)
		return intVal
	}
	return defaultValue
}

// NewApp creates a new application
func NewApp() *App {
	config := LoadConfig()

	app := &App{
		Config: config,
		Server: gin.Default(),
	}

	return app
}

// InitDatabase initializes database connection
func (a *App) InitDatabase() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		a.Config.Database.User,
		a.Config.Database.Password,
		a.Config.Database.Host,
		a.Config.Database.Port,
		a.Config.Database.Name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// Set connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	a.DB = db

	// Auto migrate tables
	if err := a.AutoMigrate(); err != nil {
		return err
	}

	log.Println("✅ Database connected successfully")
	return nil
}

// AutoMigrate runs database migrations
func (a *App) AutoMigrate() error {
	log.Println("🔄 Running database migrations...")

	err := a.DB.AutoMigrate(
		&User{},
		&RestaurantSettings{},
		&Category{},
		&MenuItem{},
		&Modifier{},
		&ModifierOption{},
		&Combo{},
		&Table{},
		&Reservation{},
		&Order{},
		&OrderItem{},
		&Payment{},
		&StockItem{},
		&StockMovement{},
		&Shift{},
		&Customer{},
		&LoyaltyTransaction{},
		&Discount{},
		&ReceiptTemplate{},
		&Printer{},
		&DailyReport{},
		&AuditLog{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✅ Database migrations completed")
	return nil
}

// SetupRoutes configures all routes
func (a *App) SetupRoutes() {
	api := a.Server.Group("/api")
	{
		// WebSocket endpoint
		a.Server.GET("/api/ws", a.WSManager.HandleWebSocket)

		// Authentication
		auth := api.Group("/auth")
		{
			auth.POST("/login", a.HandleLogin)
			auth.POST("/logout", a.HandleLogout)
			auth.POST("/refresh", a.HandleRefreshToken)
			auth.GET("/me", a.AuthMiddleware(), a.HandleGetCurrentUser)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(a.AuthMiddleware())
		{
			// Settings
			settings := protected.Group("/settings")
			{
				settings.GET("", a.HandleGetSettings)
				settings.PUT("", a.HandleUpdateSettings)
			}

			// Menu
			menu := protected.Group("/menu")
			{
				categories := menu.Group("/categories")
				{
					categories.GET("", a.HandleGetCategories)
					categories.POST("", a.HandleCreateCategory)
					categories.GET("/:id", a.HandleGetCategory)
					categories.PUT("/:id", a.HandleUpdateCategory)
					categories.DELETE("/:id", a.HandleDeleteCategory)
				}

				items := menu.Group("/items")
				{
					items.GET("", a.HandleGetMenuItems)
					items.POST("", a.HandleCreateMenuItem)
					items.GET("/:id", a.HandleGetMenuItem)
					items.PUT("/:id", a.HandleUpdateMenuItem)
					items.DELETE("/:id", a.HandleDeleteMenuItem)
				}

				modifiers := menu.Group("/modifiers")
				{
					modifiers.GET("", a.HandleGetModifiers)
					modifiers.POST("", a.HandleCreateModifier)
					modifiers.PUT("/:id", a.HandleUpdateModifier)
					modifiers.DELETE("/:id", a.HandleDeleteModifier)
				}

				combos := menu.Group("/combos")
				{
					combos.GET("", a.HandleGetCombos)
					combos.POST("", a.HandleCreateCombo)
					combos.PUT("/:id", a.HandleUpdateCombo)
					combos.DELETE("/:id", a.HandleDeleteCombo)
				}
			}

			// Orders
			orders := protected.Group("/orders")
			{
				orders.GET("", a.HandleGetOrders)
				orders.POST("", a.HandleCreateOrder)
				orders.GET("/:id", a.HandleGetOrder)
				orders.PUT("/:id", a.HandleUpdateOrder)
				orders.DELETE("/:id", a.HandleDeleteOrder)
				orders.PUT("/:id/status", a.HandleUpdateOrderStatus)
				orders.POST("/:id/items", a.HandleAddOrderItem)
				orders.PUT("/:id/items/:itemId", a.HandleUpdateOrderItem)
				orders.DELETE("/:id/items/:itemId", a.HandleDeleteOrderItem)
			}

			// Tables
			tables := protected.Group("/tables")
			{
				tables.GET("", a.HandleGetTables)
				tables.POST("", a.HandleCreateTable)
				tables.GET("/:id", a.HandleGetTable)
				tables.PUT("/:id", a.HandleUpdateTable)
				tables.DELETE("/:id", a.HandleDeleteTable)
				tables.PUT("/:id/status", a.HandleUpdateTableStatus)
				tables.PUT("/:id/transfer", a.HandleTransferTable)
			}

			// Payments
			payments := protected.Group("/payments")
			{
				payments.GET("", a.HandleGetPayments)
				payments.POST("", a.HandleCreatePayment)
				payments.POST("/refund", a.HandleRefundPayment)
			}

			// Reports
			reports := protected.Group("/reports")
			{
				reports.GET("/daily", a.HandleDailyReport)
				reports.GET("/weekly", a.HandleWeeklyReport)
				reports.GET("/monthly", a.HandleMonthlyReport)
				reports.GET("/items", a.HandleItemsReport)
				reports.GET("/categories", a.HandleCategoriesReport)
				reports.GET("/payments", a.HandlePaymentsReport)
				reports.GET("/staff", a.HandleStaffReport)
				reports.GET("/export", a.HandleExportReport)
			}

			// Inventory
			inventory := protected.Group("/inventory")
			{
				inventory.GET("/items", a.HandleGetStockItems)
				inventory.POST("/items", a.HandleCreateStockItem)
				inventory.PUT("/items/:id", a.HandleUpdateStockItem)
				inventory.DELETE("/items/:id", a.HandleDeleteStockItem)
				inventory.GET("/movements", a.HandleGetStockMovements)
				inventory.POST("/movements", a.HandleAddStockMovement)
				inventory.GET("/alerts", a.HandleGetLowStockAlerts)
			}

			// Staff
			staff := protected.Group("/staff")
			{
				staff.GET("", a.HandleGetStaff)
				staff.POST("", a.HandleCreateStaff)
				staff.PUT("/:id", a.HandleUpdateStaff)
				staff.DELETE("/:id", a.HandleDeleteStaff)
				staff.POST("/:id/shift/start", a.HandleStartShift)
				staff.POST("/:id/shift/end", a.HandleEndShift)
				staff.GET("/shifts", a.HandleGetShifts)
			}

			// Customers
			customers := protected.Group("/customers")
			{
				customers.GET("", a.HandleGetCustomers)
				customers.POST("", a.HandleCreateCustomer)
				customers.GET("/:id", a.HandleGetCustomer)
				customers.PUT("/:id", a.HandleUpdateCustomer)
				customers.DELETE("/:id", a.HandleDeleteCustomer)
				customers.POST("/:id/points", a.HandleAddLoyaltyPoints)
				customers.GET("/:id/history", a.HandleGetCustomerHistory)
			}

			// Reservations
			reservations := protected.Group("/reservations")
			{
				reservations.GET("", a.HandleGetReservations)
				reservations.POST("", a.HandleCreateReservation)
				reservations.GET("/:id", a.HandleGetReservation)
				reservations.PUT("/:id", a.HandleUpdateReservation)
				reservations.DELETE("/:id", a.HandleDeleteReservation)
				reservations.PUT("/:id/status", a.HandleUpdateReservationStatus)
			}

			// Discounts
			discounts := protected.Group("/discounts")
			{
				discounts.GET("", a.HandleGetDiscounts)
				discounts.POST("", a.HandleCreateDiscount)
				discounts.PUT("/:id", a.HandleUpdateDiscount)
				discounts.DELETE("/:id", a.HandleDeleteDiscount)
				discounts.POST("/:id/activate", a.HandleActivateDiscount)
				discounts.POST("/:id/deactivate", a.HandleDeactivateDiscount)
			}

			// Printers
			printers := protected.Group("/printers")
			{
				printers.GET("", a.HandleGetPrinters)
				printers.POST("", a.HandleCreatePrinter)
				printers.PUT("/:id", a.HandleUpdatePrinter)
				printers.DELETE("/:id", a.HandleDeletePrinter)
				printers.POST("/:id/test", a.HandleTestPrinter)
			}

			// Receipt Templates
			templates := protected.Group("/receipt-templates")
			{
				templates.GET("", a.HandleGetReceiptTemplates)
				templates.POST("", a.HandleCreateReceiptTemplate)
				templates.PUT("/:id", a.HandleUpdateReceiptTemplate)
				templates.DELETE("/:id", a.HandleDeleteReceiptTemplate)
				templates.POST("/:id/set-default", a.HandleSetDefaultReceiptTemplate)
			}

			// Print Actions
			print := protected.Group("/print")
			{
				print.POST("/receipt/:orderId", a.HandlePrintReceipt)
				print.POST("/kitchen/:orderId", a.HandlePrintKitchen)
				print.POST("/bar/:orderId", a.HandlePrintBar)
			}

			// WhatsApp
			whatsapp := protected.Group("/whatsapp")
			{
				whatsapp.POST("/receipt/:orderId", a.HandleSendWhatsAppReceipt)
				whatsapp.POST("/daily-report", a.HandleSendWhatsAppDailyReport)
				whatsapp.POST("/test", a.HandleTestWhatsApp)
			}

			// Dashboard
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("", a.HandleGetDashboard)
				dashboard.GET("/stats", a.HandleGetDashboardStats)
				dashboard.GET("/recent-orders", a.HandleGetRecentOrders)
				dashboard.GET("/low-stock", a.HandleGetLowStockDashboard)
			}
		}
	}

	// WebSocket (if needed)
	// a.Server.GET("/ws", a.HandleWebSocket)

	// Serve static files (frontend)
	a.Server.Static("/static", "../frontend/dist")
	a.Server.StaticFile("/", "../frontend/dist/index.html")

	// Health check
	a.Server.GET("/health", a.HandleHealthCheck)
}

// Start starts to server
func (a *App) Start() error {
	addr := fmt.Sprintf("%s:%d", a.Config.Server.Host, a.Config.Server.Port)
	log.Printf("🚀 Server starting on %s", addr)
	log.Printf("📊 Database: %s", a.Config.Database.Name)
	log.Printf("🌐 JWT Expiration: %v", a.Config.JWT.Expiration)
	log.Printf("📱 WhatsApp: %v", a.Config.WhatsApp.Enabled)
	log.Printf("📧 Email: %v", a.Config.Email.Enabled)

	return a.Server.Run(addr)
}

// HandleHealthCheck handles health check requests
func (a *App) HandleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"version": "1.0.0",
		"time":     time.Now().Format(time.RFC3339),
		"websocket": "enabled",
	})
}

// AuthMiddleware validates JWT tokens
func (a *App) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// Validate token
		claims, err := a.ValidateJWTToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// ValidateJWTToken validates a JWT token
func (a *App) ValidateJWTToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(a.Config.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Extract claims
	if claims, ok := token.Claims.(*JWTClaims); ok && claims.Valid(time.Now()) {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GenerateJWTToken generates a JWT token
func (a *App) GenerateJWTToken(userID uint, email, role string) (string, error) {
	claims := &JWTClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		ExpiresAt: time.Now().Add(a.Config.JWT.Expiration),
		IssuedAt:  time.Now(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.Config.JWT.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserIDFromContext extracts user ID from context
func (a *App) GetUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if idStr, ok := userID.(uint); ok {
			return fmt.Sprintf("%d", idStr)
		}
	}
	return ""
}

// Main entry point
func main() {
	app := NewApp()

	// Create WebSocket manager
	wsManager := NewWebSocketManager()

	// Create notification service
	notificationService := NewNotificationService(wsManager, app.DB)

	// Store in app
	app.WSManager = wsManager
	app.NotificationService = notificationService

	// Initialize database
	if err := app.InitDatabase(); err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}

	// Start WebSocket manager in goroutine
	go wsManager.Run()

	// Start periodic dashboard updates (every 30 seconds)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if app.DB != nil {
				stats, err := app.GetDashboardStatsInternal()
				if err == nil {
					// Broadcast to all dashboard clients
					wsManager.CreateDashboardUpdateNotification(stats)
				}
			}
		}
	}()

	// Setup routes
	app.SetupRoutes()

	// Seed data (optional)
	app.SeedData()

	// Start server
	if err := app.Start(); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// getDashboardStatsInternal gets dashboard stats (internal helper)
func (a *App) GetDashboardStatsInternal() (DashboardStats, error) {
	stats := DashboardStats{}

	// Get date range (default today)
	startDate := time.Now().Truncate(24 * time.Hour)
	endDate := startDate.Add(24 * time.Hour)

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

	return stats, nil
}
