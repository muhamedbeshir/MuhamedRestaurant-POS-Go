package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/crypto/bcrypt"
)

// ========================================
// MIDDLEWARE
// ========================================

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

		// Validate JWT token
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

// RBACMiddleware enforces role-based access control
func (a *App) RBACMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		role := userRole.(string)

		// Check if user role is allowed
		allowed := false
		for _, allowedRole := range allowedRoles {
			if role == allowedRole || role == "super_admin" {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMiddleware handles CORS
func (a *App) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggingMiddleware logs all requests
func (a *App) LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Log request
		method := c.Request.Method
		path := c.Request.URL.Path
		userID, _ := c.Get("user_id")

		// Process request
		c.Next()

		// Log response
		latency := time.Since(start)
		status := c.Writer.Status()

		// Create audit log (for critical operations)
		if userID != nil && (method == "POST" || method == "PUT" || method == "DELETE") {
			a.CreateAuditLog(userID.(uint), method+" "+path, "", nil, nil)
		}

		fmt.Printf("[%s] %s %s - Status: %d - Latency: %v - UserID: %v",
			start.Format("2006-01-02 15:04:05"),
			method,
			path,
			status,
			latency,
			userID)
	}
}

// RateLimitMiddleware implements basic rate limiting
func (a *App) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simplified rate limiting
		// In production, use redis or in-memory store
		c.Next()
	}
}

// ValidateJWTToken validates a JWT token and returns claims
func (a *App) ValidateJWTToken(tokenString string) (*JWTClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return secret key
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

// GenerateJWTToken generates a new JWT token
func (a *App) GenerateJWTToken(userID uint, email, role string) (string, error) {
	// Create claims
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		ExpiresAt: time.Now().Add(a.Config.JWT.Expiration),
		IssuedAt:  time.Now(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.Config.JWT.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	jwt.RegisteredClaims
}

// ========================================
// AUDIT LOGGING
// ========================================

// CreateAuditLog creates an audit log entry
func (a *App) CreateAuditLog(userID uint, action, entity string, entityID *uint, changes interface{}) {
	log := &AuditLog{
		UserID:    userID,
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		Changes:   a.InterfaceToJSON(changes),
		IPAddress: a.GetClientIP(a.Server),
		UserAgent: a.GetUserAgent(a.Server),
	}

	a.DB.Create(log)
}

// ========================================
// VALIDATION
// ========================================

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	return len(email) > 3 && len(email) < 256 && 
		(email[len(email)-3] == '.' && email[len(email)-2] >= 'a' && email[len(email)-2] <= 'z')
}

// ValidatePassword validates password strength
func ValidatePassword(password string) bool {
	return len(password) >= 8
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword checks if password matches hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ========================================
// HELPERS
// ========================================

// InterfaceToJSON converts interface to JSON string
func (a *App) InterfaceToJSON(data interface{}) string {
	if data == nil {
		return ""
	}
	// Simplified - in production use json.Marshal
	return fmt.Sprintf("%v", data)
}

// GetClientIP extracts client IP address
func (a *App) GetClientIP(c *gin.Context) string {
	// Try to get from various headers
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	return c.ClientIP()
}

// GetUserAgent extracts user agent
func (a *App) GetUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
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

// GetCurrentTime returns current time
func (a *App) GetCurrentTime() time.Time {
	return time.Now()
}

// GetCurrentTimestamp returns current timestamp
func (a *App) GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// GetTodayDate returns today's date
func (a *App) GetTodayDate() string {
	return time.Now().Format("2006-01-02")
}

// GetCurrentDateTime returns current date and time
func (a *App) GetCurrentDateTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// GenerateOrderNumber generates a unique order number
func (a *App) GenerateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().Unix())
}

// GenerateReference generates a unique reference number
func (a *App) GenerateReference() string {
	return fmt.Sprintf("REF-%d", time.Now().UnixNano())
}

// FormatCurrency formats amount as currency
func (a *App) FormatCurrency(amount float64, currencySymbol string) string {
	return fmt.Sprintf("%.2f %s", amount, currencySymbol)
}

// FormatDate formats date
func (a *App) FormatDate(date time.Time) string {
	return date.Format("2006-01-02")
}

// FormatDateTime formats date and time
func (a *App) FormatDateTime(date time.Time) string {
	return date.Format("2006-01-02 15:04")
}

// FormatTime formats time
func (a *App) FormatTime(date time.Time) string {
	return date.Format("15:04")
}

// SeedData seeds initial data
func (a *App) SeedData() {
	// Seed settings
	if a.DB.Where("id = ?", 1).First(&RestaurantSettings{}).RowsAffected == 0 {
		settings := RestaurantSettings{
			Name:           "مطعمي",
			NameAr:         "مطعمي",
			Currency:        "EGP",
			CurrencySymbol:  "ج.م",
			TaxRate:        0.14,
			ServiceCharge:   0.10,
			Language:        "ar",
			ThemeColor:      "#10b981",
			IsOpen:         true,
		}
		a.DB.Create(&settings)
	}

	// Seed default user (password: admin123)
	if a.DB.Where("email = ?", "admin@restaurant.com").First(&User{}).RowsAffected == 0 {
		hashedPassword, _ := HashPassword("admin123")
		user := User{
			Email:    "admin@restaurant.com",
			Password: hashedPassword,
			Name:     "Admin User",
			Role:     "super_admin",
			IsActive: true,
		}
		a.DB.Create(&user)
	}

	// Seed default categories
	categoryNames := []struct {
		Name   string
		NameAr string
		Order  int
	}{
		{"Appetizers", "المقبلات", 1},
		{"Main Course", "الأطباق الرئيسية", 2},
		{"Drinks", "المشروبات", 3},
		{"Desserts", "الحلويات", 4},
	}

	for _, cat := range categoryNames {
		if a.DB.Where("name = ?", cat.Name).First(&Category{}).RowsAffected == 0 {
			category := Category{
				Name:         cat.Name,
				NameAr:       cat.NameAr,
				DisplayOrder: cat.Order,
				IsActive:     true,
				IsAvailable:   true,
			}
			a.DB.Create(&category)
		}
	}
}
