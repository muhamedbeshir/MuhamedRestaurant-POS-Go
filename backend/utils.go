package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ========================================
// UTILS
// ========================================

// DBUtils handles database utilities
type DBUtils struct {
	DB *gorm.DB
}

// BackupDatabase creates a database backup
func (d *DBUtils) BackupDatabase(filename string) error {
	// Get database connection
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}

	// Use mysqldump to create backup
	backupFile := "/tmp/" + filename + ".sql"
	cmd := exec.Command("mysqldump", "--opt", "--all", "--add-drop-database", "-u", "root", "restaurant_pos")

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	return os.WriteFile(backupFile, output, 0644)
}

// RestoreDatabase restores database from backup
func (d *DBUtils) RestoreDatabase(filename string) error {
	backupFile := "/tmp/" + filename + ".sql"

	// Read backup file
	content, err := os.ReadFile(backupFile)
	if err != nil {
		return err
	}

	// Execute SQL
	_, err = d.DB.Exec(string(content))
	return err
}

// ========================================
// FILE UTILS
// ========================================

// SaveToFile saves content to a file
func SaveToFile(filename, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filename, []byte(content), 0644)
}

// ReadFromFile reads content from a file
func ReadFromFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	return string(content), err
}

// DeleteFile deletes a file
func DeleteFile(filename string) error {
	return os.Remove(filename)
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// ========================================
// DATE/TIME UTILS
// ========================================

// ParseDate parses date string
func ParseDate(dateStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02",
		"2006/01/02",
		"01-02-2006",
		"01/02/2006",
	}

	for _, layout := range layouts {
		if date, err := time.Parse(layout, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
}

// GetDateRange returns date range string
func GetDateRange(startDate, endDate time.Time) string {
	return fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
}

// GetTimeAgo returns time ago string
func GetTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return fmt.Sprintf("%d seconds ago", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	} else if duration < 30*24*time.Hour {
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	} else if duration < 12*30*24*time.Hour {
		return fmt.Sprintf("%d months ago", int(duration.Hours()/(24*30)))
	} else {
		return fmt.Sprintf("%d years ago", int(duration.Hours()/(24*30*12)))
	}
}

// ========================================
// STRING UTILS
// ========================================

// TruncateString truncates string to max length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// SanitizeString removes dangerous characters
func SanitizeString(s string) string {
	// Remove SQL injection attempts
	s = strings.ReplaceAll(s, "'", "''")
	s = strings.ReplaceAll(s, ";", "")
	s = strings.ReplaceAll(s, "--", "")
	// Remove XSS attempts
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// Slugify converts string to URL-friendly slug
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove special characters
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, s)

	return s
}

// ========================================
// NUMBER UTILS
// ========================================

// FormatNumber formats number with thousands separator
func FormatNumber(num float64) string {
	// Simplified version
	return fmt.Sprintf("%.2f", num)
}

// RoundToDecimal rounds to 2 decimal places
func RoundToDecimal(num float64) float64 {
	rounded, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", num), 64)
	return rounded
}

// CalculatePercentage calculates percentage
func CalculatePercentage(part, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (part / total) * 100
}

// ========================================
// VALIDATION UTILS
// ========================================

// IsValidEmail validates email
func IsValidEmail(email string) bool {
	return len(email) > 3 && len(email) < 256 &&
		strings.Contains(email, "@") &&
		strings.Contains(email, ".") &&
		strings.LastIndex(email, "@") < strings.LastIndex(email, ".")
}

// IsValidPhone validates phone number
func IsValidPhone(phone string) bool {
	// Simplified validation for Egyptian numbers
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "+", "")
	return len(phone) == 11 && strings.HasPrefix(phone, "01")
}

// IsValidURL validates URL
func IsValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://")
}

// ========================================
// ENCRYPTION UTILS
// ========================================

// Encrypt encrypts string (simplified)
func Encrypt(plaintext, key string) (string, error) {
	// In production, use proper encryption (AES, etc.)
	// This is a simplified version
	return base64.StdEncoding.EncodeToString([]byte(plaintext)), nil
}

// Decrypt decrypts string (simplified)
func Decrypt(ciphertext, key string) (string, error) {
	// In production, use proper decryption
	// This is a simplified version
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ========================================
// JSON UTILS
// ========================================

// ToJSON converts struct to JSON string
func ToJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	return string(data), err
}

// FromJSON converts JSON string to struct
func FromJSON(jsonStr string, v interface{}) error {
	return json.Unmarshal([]byte(jsonStr), v)
}

// ========================================
// HTTP UTILS
// ========================================

// SendHTTPRequest sends HTTP request
func SendHTTPRequest(method, url string, body interface{}, headers map[string]string) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	return client.Do(req)
}

// ========================================
// FILE UPLOAD UTILS
// ========================================

// SaveUploadedFile saves uploaded file
func SaveUploadedFile(file io.Reader, filename string) (string, error) {
	// Create uploads directory
	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return "", err
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)
	uniqueFilename := fmt.Sprintf("%s_%d%s", nameWithoutExt, time.Now().UnixNano(), ext)

	// Save file
	filePath := filepath.Join(uploadsDir, uniqueFilename)
	fileHandle, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer fileHandle.Close()

	_, err = io.Copy(fileHandle, file)
	return filePath, err
}

// ========================================
// WHATSAPP API UTILS
// ========================================

// WhatsAppAPIService handles WhatsApp API
type WhatsAppAPIService struct {
	APIURL  string
	APIKey  string
	Enabled bool
}

// SendWhatsAppAPIMessage sends message via WhatsApp API
func (w *WhatsAppAPIService) SendWhatsAppAPIMessage(phoneNumber, message string) error {
	if !w.Enabled {
		return fmt.Errorf("WhatsApp is disabled")
	}

	// Example using Twilio WhatsApp API
	// Replace with actual WhatsApp API (Twilio, MessageBird, etc.)
	
	url := fmt.Sprintf("%s/messages", w.APIURL)

	payload := map[string]interface{}{
		"from": "whatsapp:+14155238886",
		"to":   "whatsapp:" + phoneNumber,
		"body": message,
	}

	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("WhatsApp API returned status %d", resp.StatusCode)
	}

	return nil
}

// SendWhatsAppImage sends image via WhatsApp
func (w *WhatsAppAPIService) SendWhatsAppImage(phoneNumber, mediaURL, caption string) error {
	if !w.Enabled {
		return fmt.Errorf("WhatsApp is disabled")
	}

	url := fmt.Sprintf("%s/messages", w.APIURL)

	payload := map[string]interface{}{
		"from": "whatsapp:+14155238886",
		"to":   "whatsapp:" + phoneNumber,
		"type": "image",
		"media": map[string]interface{}{
			"url": mediaURL,
		},
		"caption": caption,
	}

	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("WhatsApp API returned status %d", resp.StatusCode)
	}

	return nil
}

// SendWhatsAppDocument sends document via WhatsApp
func (w *WhatsAppAPIService) SendWhatsAppDocument(phoneNumber, mediaURL, filename string) error {
	if !w.Enabled {
		return fmt.Errorf("WhatsApp is disabled")
	}

	url := fmt.Sprintf("%s/messages", w.APIURL)

	payload := map[string]interface{}{
		"from": "whatsapp:+14155238886",
		"to":   "whatsapp:" + phoneNumber,
		"type": "document",
		"media": map[string]interface{}{
			"url":  mediaURL,
			"filename": filename,
		},
	}

	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("WhatsApp API returned status %d", resp.StatusCode)
	}

	return nil
}

// ========================================
// CONFIG UTILS
// ========================================

// GetConfigValue gets config value from database
func GetConfigValue(db *gorm.DB, key string) (string, error) {
	var value string
	err := db.Table("settings").Where("`key` = ?", key).Select("value").First(&value).Error
	return value, err
}

// SetConfigValue sets config value in database
func SetConfigValue(db *gorm.DB, key, value string) error {
	return db.Table("settings").Where("`key` = ?", key).Update("value", value).Error
}

// GetEnv gets environment variable or default
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt gets environment variable as int or default
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intVal int
		fmt.Sscanf(value, "%d", &intVal)
		return intVal
	}
	return defaultValue
}

// GetEnvBool gets environment variable as bool or default
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

// ========================================
// EXPORT/IMPORT UTILS
// ========================================

// ExportToCSV exports data to CSV
func ExportToCSV(data []map[string]interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if len(data) == 0 {
		return nil
	}

	// Write headers
	headers := make([]string, 0)
	for k := range data[0] {
		headers = append(headers, k)
	}
	file.WriteString(strings.Join(headers, ",") + "\n")

	// Write data
	for _, row := range data {
		values := make([]string, 0)
		for _, h := range headers {
			values = append(values, fmt.Sprintf("%v", row[h]))
		}
		file.WriteString(strings.Join(values, ",") + "\n")
	}

	return nil
}

// ExportToJSON exports data to JSON
func ExportToJSON(data interface{}, filename string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}

// ImportFromJSON imports data from JSON
func ImportFromJSON(filename string, v interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// ========================================
// RESPONSE UTILS
// ========================================

// SuccessResponse creates success response
func SuccessResponse(message string, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"message": message,
		"data":    data,
	}
}

// ErrorResponse creates error response
func ErrorResponse(message, details string) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"error":   message,
		"details": details,
	}
}

// PaginationResponse creates paginated response
func PaginationResponse(data interface{}, page, limit int, total int64) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"data":    data,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	}
}

// ========================================
// LOGGING UTILS
// ========================================

// LogInfo logs info message
func LogInfo(message string) {
	fmt.Printf("[INFO] %s - %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

// LogError logs error message
func LogError(message string, err error) {
	fmt.Printf("[ERROR] %s - %s: %v\n", time.Now().Format("2006-01-02 15:04:05"), message, err)
}

// LogWarning logs warning message
func LogWarning(message string) {
	fmt.Printf("[WARNING] %s - %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

// LogDebug logs debug message
func LogDebug(message string) {
	if os.Getenv("DEBUG") == "true" {
		fmt.Printf("[DEBUG] %s - %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
	}
}
