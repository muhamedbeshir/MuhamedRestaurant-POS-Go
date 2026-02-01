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
	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// App structure
type App struct {
	DB     *gorm.DB
	Server *gin.Engine
	Config *Config
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
		APIURL    string
		APIKey    string
		Enabled    bool
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
			APIURL string
			APIKey string
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

// InitDatabase initializes the database connection
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

// Start starts the server
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
		"status":  "ok",
		"version": "1.0.0",
		"time":     time.Now().Format(time.RFC3339),
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

		// Validate token (simplified - in production use proper JWT library)
		userID, err := a.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

// ValidateToken validates a JWT token (simplified)
func (a *App) ValidateToken(token string) (string, error) {
	// In production, use proper JWT library like dgrijalva/jwt-go
	// For now, return a mock user ID
	return "1", nil
}

// GenerateToken generates a JWT token (simplified)
func (a *App) GenerateToken(userID string) (string, error) {
	// In production, use proper JWT library
	// For now, return a mock token
	return "mock-jwt-token-" + userID, nil
}

// GetUserIDFromContext extracts user ID from context
func (a *App) GetUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if idStr, ok := userID.(string); ok {
			return idStr
		}
	}
	return ""
}

// Main entry point
func main() {
	app := NewApp()

	// Initialize database
	if err := app.InitDatabase(); err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}

	// Setup routes
	app.SetupRoutes()

	// Seed data (optional)
	app.SeedData()

	// Start server
	if err := app.Start(); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
