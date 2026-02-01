package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ========================================
// SERVICES
// ========================================

// OrderService handles order business logic
type OrderService struct {
	DB *gorm.DB
}

// CreateOrderWithCalculations creates order with auto calculations
func (s *OrderService) CreateOrderWithCalculations(order *Order) error {
	// Calculate totals
	var subtotal float64
	for _, item := range order.Items {
		subtotal += float64(item.Quantity) * item.UnitPrice
	}

	// Get settings for tax rate
	var settings RestaurantSettings
	s.DB.First(&settings)

	taxAmount := subtotal * settings.TaxRate
	serviceCharge := subtotal * settings.ServiceCharge
	total := subtotal + taxAmount + serviceCharge

	// Update order
	order.Subtotal = subtotal
	order.TaxAmount = taxAmount
	order.ServiceCharge = serviceCharge
	order.Total = total
	order.OrderNumber = fmt.Sprintf("ORD-%d", time.Now().Unix())

	// Save order
	return s.DB.Create(order).Error
}

// WhatsAppService handles WhatsApp messaging
type WhatsAppService struct {
	APIURL  string
	APIKey  string
	Enabled bool
}

// SendWhatsAppMessage sends message via WhatsApp API
func (w *WhatsAppService) SendWhatsAppMessage(phoneNumber, message string) error {
	if !w.Enabled {
		return fmt.Errorf("WhatsApp is disabled")
	}

	// Example using Twilio or any WhatsApp API
	// Replace with actual WhatsApp API integration
	
	// Example payload
	payload := map[string]interface{}{
		"to":      phoneNumber,
		"message": message,
	}

	jsonPayload, _ := json.Marshal(payload)

	// Send HTTP request
	req, _ := http.NewRequest("POST", w.APIURL, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WhatsApp API returned status %d", resp.StatusCode)
	}

	return nil
}

// SendWhatsAppReceipt sends order receipt via WhatsApp
func (w *WhatsAppService) SendWhatsAppReceipt(order *Order, settings *RestaurantSettings) error {
	// Format receipt message
	message := w.formatReceiptMessage(order, settings)
	
	// Get customer phone from order
	phone := order.CustomerPhone
	if phone == "" {
		return fmt.Errorf("no customer phone number")
	}

	// Send message
	return w.SendWhatsAppMessage(phone, message)
}

// formatReceiptMessage formats receipt for WhatsApp
func (w *WhatsAppService) formatReceiptMessage(order *Order, settings *RestaurantSettings) string {
	var sb strings.Builder

	// Restaurant name
	sb.WriteString(fmt.Sprintf("*%s*\n\n", settings.Name))

	// Order number
	sb.WriteString(fmt.Sprintf("📋 Order #%s\n", order.OrderNumber))
	sb.WriteString(fmt.Sprintf("📅 %s\n\n", order.CreatedAt.Format("2006-01-02 15:04")))

	// Items
	sb.WriteString("*الطلب:*\n")
	for _, item := range order.Items {
		sb.WriteString(fmt.Sprintf("• %s x%d\n", item.MenuItemName, item.Quantity))
		sb.WriteString(fmt.Sprintf("  %s%.2f\n", settings.CurrencySymbol, item.UnitPrice))
	}

	sb.WriteString("\n")

	// Totals
	sb.WriteString(fmt.Sprintf("المجموع: %s%.2f\n", settings.CurrencySymbol, order.Subtotal))
	sb.WriteString(fmt.Sprintf("الضريبة: %s%.2f\n", settings.CurrencySymbol, order.TaxAmount))
	sb.WriteString(fmt.Sprintf("*الإجمالي: %s%.2f*\n", settings.CurrencySymbol, order.Total))

	sb.WriteString("\n")
	sb.WriteString("شكراً لاختيارك! 🙏")

	return sb.String()
}

// SendWhatsAppDailyReport sends daily sales report via WhatsApp
func (w *WhatsAppService) SendWhatsAppDailyReport(report *DailyReport, settings *RestaurantSettings, phone string) error {
	// Format daily report message
	message := w.formatDailyReportMessage(report, settings)

	// Send message
	return w.SendWhatsAppMessage(phone, message)
}

// formatDailyReportMessage formats daily report for WhatsApp
func (w *WhatsAppService) formatDailyReportMessage(report *DailyReport, settings *RestaurantSettings) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("*📊 تقرير يومي - %s*\n\n", report.ReportDate.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("📈 إجمالي الإيرادات: %s%.2f\n", settings.CurrencySymbol, report.TotalRevenue))
	sb.WriteString(fmt.Sprintf("📦 عدد الطلبات: %d\n", report.TotalOrders))
	sb.WriteString(fmt.Sprintf("🍽️ طلبات صالون: %d\n", report.DineInOrders))
	sb.WriteString(fmt.Sprintf("📦 طلبات تاك أواي: %d\n", report.TakeawayOrders))
	sb.WriteString(fmt.Sprintf("🚗 طلبات دليفري: %d\n", report.DeliveryOrders))

	sb.WriteString("\n")
	sb.WriteString("💰 المدفوعات:\n")
	sb.WriteString(fmt.Sprintf("  كاش: %s%.2f\n", settings.CurrencySymbol, report.CashPayments))
	sb.WriteString(fmt.Sprintf("  بطاقة: %s%.2f\n", settings.CurrencySymbol, report.CardPayments))
	sb.WriteString(fmt.Sprintf("  محفظة: %s%.2f\n", settings.CurrencySymbol, report.WalletPayments))

	return sb.String()
}

// EmailService handles email sending
type EmailService struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	Enabled       bool
}

// SendEmail sends an email
func (e *EmailService) SendEmail(to, subject, body string) error {
	if !e.Enabled {
		return fmt.Errorf("Email is disabled")
	}

	// Use sendmail or any email library
	// This is a simplified version
	// In production, use gomail or similar library

	cmd := exec.Command("sendmail", "-t")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("To: %s\nSubject: %s\n\n%s", to, subject, body))
	
	return cmd.Run()
}

// SendReceiptByEmail sends order receipt via email
func (e *EmailService) SendReceiptByEmail(order *Order, settings *RestaurantSettings, email string) error {
	subject := fmt.Sprintf("فاتورة - %s", order.OrderNumber)
	body := e.formatReceiptEmail(order, settings)

	return e.SendEmail(email, subject, body)
}

// formatReceiptEmail formats receipt for email
func (e *EmailService) formatReceiptEmail(order *Order, settings *RestaurantSettings) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("<h1>%s</h1>", settings.Name))
	sb.WriteString(fmt.Sprintf("<p>Order #%s</p>", order.OrderNumber))
	sb.WriteString(fmt.Sprintf("<p>Date: %s</p>", order.CreatedAt.Format("2006-01-02 15:04")))

	sb.WriteString("<h3>Items:</h3><ul>")
	for _, item := range order.Items {
		sb.WriteString(fmt.Sprintf("<li>%s x%d - %s%.2f</li>",
			item.MenuItemName, item.Quantity, settings.CurrencySymbol, item.UnitPrice))
	}
	sb.WriteString("</ul>")

	sb.WriteString("<h3>Totals:</h3>")
	sb.WriteString(fmt.Sprintf("<p>Subtotal: %s%.2f</p>", settings.CurrencySymbol, order.Subtotal))
	sb.WriteString(fmt.Sprintf("<p>Tax: %s%.2f</p>", settings.CurrencySymbol, order.TaxAmount))
	sb.WriteString(fmt.Sprintf("<p><strong>Total: %s%.2f</strong></p>", settings.CurrencySymbol, order.Total))

	return sb.String()
}

// PrintService handles printing
type PrintService struct {
	Settings *RestaurantSettings
	Runtime  *runtime.Runtime
}

// PrintReceipt prints receipt directly to printer (no dialog)
func (p *PrintService) PrintReceipt(order *Order, printer string) error {
	// Generate receipt HTML
	receiptHTML := p.generateReceiptHTML(order)

	// Save to temporary file
	tempFile := "/tmp/receipt_" + order.OrderNumber + ".html"
	if err := os.WriteFile(tempFile, []byte(receiptHTML), 0644); err != nil {
		return err
	}

	// Print using system print command (direct, no dialog)
	// Linux: lp command
	// Windows: powershell or lpr
	// Mac: lp or lpr
	
	var cmd *exec.Cmd
	if strings.Contains(runtime.GOOS, "linux") {
		// Linux
		if printer != "" {
			cmd = exec.Command("lp", "-d", printer, tempFile)
		} else {
			cmd = exec.Command("lp", tempFile)
		}
	} else if strings.Contains(runtime.GOOS, "windows") {
		// Windows
		if printer != "" {
			cmd = exec.Command("powershell", "-Command",
				fmt.Sprintf("Get-Content %s | Out-Printer -Name %s", tempFile, printer))
		} else {
			cmd = exec.Command("powershell", "-Command",
				fmt.Sprintf("Get-Content %s | Out-Printer", tempFile))
		}
	} else if strings.Contains(runtime.GOOS, "darwin") {
		// Mac
		if printer != "" {
			cmd = exec.Command("lp", "-d", printer, tempFile)
		} else {
			cmd = exec.Command("lp", tempFile)
		}
	} else {
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	return cmd.Run()
}

// PrintKitchen prints kitchen ticket
func (p *PrintService) PrintKitchen(order *Order, printer string) error {
	// Generate kitchen HTML
	kitchenHTML := p.generateKitchenHTML(order)

	// Save to temporary file
	tempFile := "/tmp/kitchen_" + order.OrderNumber + ".html"
	if err := os.WriteFile(tempFile, []byte(kitchenHTML), 0644); err != nil {
		return err
	}

	// Print
	var cmd *exec.Cmd
	if strings.Contains(runtime.GOOS, "linux") {
		cmd = exec.Command("lp", "-d", printer, tempFile)
	} else if strings.Contains(runtime.GOOS, "windows") {
		cmd = exec.Command("powershell", "-Command",
			fmt.Sprintf("Get-Content %s | Out-Printer -Name %s", tempFile, printer))
	} else {
		cmd = exec.Command("lp", "-d", printer, tempFile)
	}

	return cmd.Run()
}

// generateReceiptHTML generates HTML receipt
func (p *PrintService) generateReceiptHTML(order *Order) string {
	direction := "rtl"
	if p.Settings.Language == "en" {
		direction = "ltr"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html dir="%s">
<head>
    <meta charset="UTF-8">
    <title>Receipt</title>
    <style>
        body { font-family: 'Cairo', Arial, sans-serif; width: 80mm; margin: 0; padding: 10px; }
        .header { text-align: center; margin-bottom: 10px; }
        .header h1 { font-size: 18px; margin: 5px 0; }
        .line { border-bottom: 1px dashed #000; margin: 10px 0; }
        .item { display: flex; justify-content: space-between; margin: 5px 0; }
        .total { text-align: right; font-weight: bold; font-size: 16px; margin-top: 15px; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>%s</h1>
        <p>%s</p>
    </div>
    <div class="line"></div>`, direction, p.Settings.Name, order.OrderNumber))

	// Items
	for _, item := range order.Items {
		sb.WriteString(fmt.Sprintf(`<div class="item">
            <span>%s x%d</span>
            <span>%.2f</span>
        </div>`, item.MenuItemName, item.Quantity, item.UnitPrice))
	}

	sb.WriteString(fmt.Sprintf(`    <div class="line"></div>
    <div class="total">Total: %.2f</div>
    <div class="footer">
        <p>%s</p>
        <p>%s</p>
    </div>
</body>
</html>`, order.Total, p.Settings.Name, p.Settings.ReceiptFooter))

	return sb.String()
}

// generateKitchenHTML generates kitchen ticket
func (p *PrintService) generateKitchenHTML(order *Order) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html dir="ltr">
<head>
    <meta charset="UTF-8">
    <title>Kitchen Order</title>
    <style>
        body { font-family: Arial, sans-serif; width: 80mm; margin: 0; padding: 10px; }
        .header { text-align: center; margin-bottom: 10px; }
        .order-number { font-size: 20px; font-weight: bold; }
        .line { border-bottom: 1px dashed #000; margin: 10px 0; }
        .item { margin: 10px 0; }
        .item-name { font-weight: bold; font-size: 14px; }
        .item-qty { font-size: 18px; font-weight: bold; }
        .urgent { background: #ff0000; color: white; }
    </style>
</head>
<body>
    <div class="header">
        <p>KITCHEN ORDER</p>
        <p class="order-number">#%s</p>
    </div>
    <div class="line"></div>`)

	// Items
	for _, item := range order.Items {
		sb.WriteString(fmt.Sprintf(`<div class="item">
        <div class="item-name">%s</div>
        <div class="item-qty">x%d</div>
    </div>`, item.MenuItemName, item.Quantity))
	}

	sb.WriteString(`    <div class="line"></div>
</body>
</html>`)

	return sb.String()
}

// ReportService handles report generation
type ReportService struct {
	DB *gorm.DB
}

// GenerateDailyReport generates daily sales report
func (r *ReportService) GenerateDailyReport(date time.Time) (*DailyReport, error) {
	var report DailyReport

	// Get date range
	startDate := date.Truncate(24 * time.Hour)
	endDate := startDate.Add(24 * time.Hour)

	// Get total orders
	r.DB.Model(&Order{}).Where("created_at >= ? AND created_at < ?", startDate, endDate).Count(&report.TotalOrders)

	// Get total revenue
	r.DB.Model(&Order{}).Where("created_at >= ? AND created_at < ? AND payment_status = ?", startDate, endDate, "paid").
		Select("COALESCE(SUM(total), 0)").Scan(&report.TotalRevenue)

	// Get orders by type
	r.DB.Model(&Order{}).Where("created_at >= ? AND created_at < ? AND type = ?", startDate, endDate, "dine_in").Count(&report.DineInOrders)
	r.DB.Model(&Order{}).Where("created_at >= ? AND created_at < ? AND type = ?", startDate, endDate, "takeaway").Count(&report.TakeawayOrders)
	r.DB.Model(&Order{}).Where("created_at >= ? AND created_at < ? AND type = ?", startDate, endDate, "delivery").Count(&report.DeliveryOrders)

	// Get payments by method
	r.DB.Model(&Payment{}).Where("created_at >= ? AND created_at < ? AND method = ?", startDate, endDate, "cash").
		Select("COALESCE(SUM(amount), 0)").Scan(&report.CashPayments)
	r.DB.Model(&Payment{}).Where("created_at >= ? AND created_at < ? AND method = ?", startDate, endDate, "card").
		Select("COALESCE(SUM(amount), 0)").Scan(&report.CardPayments)
	r.DB.Model(&Payment{}).Where("created_at >= ? AND created_at < ? AND method = ?", startDate, endDate, "mobile_wallet").
		Select("COALESCE(SUM(amount), 0)").Scan(&report.WalletPayments)

	// Get top items
	type ItemCount struct {
		Name  string
		Count int
	}
	var topItems []ItemCount
	r.DB.Model(&OrderItem{}).
		Select("menu_item_name, SUM(quantity) as count").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group("menu_item_name").
		Order("count DESC").
		Limit(10).
		Scan(&topItems)

	topItemsJSON, _ := json.Marshal(topItems)
	report.TopItems = string(topItemsJSON)

	// Save report
	report.ReportDate = date
	r.DB.Save(&report)

	return &report, nil
}

// LoyaltyService handles loyalty points
type LoyaltyService struct {
	DB *gorm.DB
}

// AddLoyaltyPoints adds loyalty points to customer
func (l *LoyaltyService) AddLoyaltyPoints(customerID uint, points int, orderID uint, notes string) error {
	transaction := &LoyaltyTransaction{
		CustomerID: customerID,
		Type:       "earned",
		Points:     points,
		OrderID:    &orderID,
		Notes:      notes,
	}

	if err := l.DB.Create(transaction).Error; err != nil {
		return err
	}

	// Update customer points
	l.DB.Model(&Customer{}).Where("id = ?", customerID).
		Update("points", gorm.Expr("points + ?", points))

	// Update total spent (if this is a purchase)
	l.DB.Model(&Customer{}).Where("id = ?", customerID).
		Update("total_spent", gorm.Expr("total_spent + ?", points))

	// Update visits count
	l.DB.Model(&Customer{}).Where("id = ?", customerID).
		UpdateColumn("visits_count", gorm.Expr("visits_count + 1"))

	return nil
}

// RedeemLoyaltyPoints redeems loyalty points
func (l *LoyaltyService) RedeemLoyaltyPoints(customerID uint, points int, notes string) error {
	// Check if customer has enough points
	var customer Customer
	if err := l.DB.First(&customer, customerID).Error; err != nil {
		return err
	}

	if customer.Points < points {
		return fmt.Errorf("not enough loyalty points")
	}

	transaction := &LoyaltyTransaction{
		CustomerID: customerID,
		Type:       "redeemed",
		Points:     -points,
		Notes:      notes,
	}

	if err := l.DB.Create(transaction).Error; err != nil {
		return err
	}

	// Update customer points
	l.DB.Model(&Customer{}).Where("id = ?", customerID).
		Update("points", gorm.Expr("points - ?", points))

	return nil
}

// InventoryService handles inventory
type InventoryService struct {
	DB *gorm.DB
}

// AddStockMovement adds a stock movement
func (i *InventoryService) AddStockMovement(stockItemID uint, movementType string, quantity float64, costPerUnit float64, reason, reference string) error {
	movement := &StockMovement{
		StockItemID: stockItemID,
		Type:         movementType,
		Quantity:     quantity,
		CostPerUnit:  costPerUnit,
		Reason:       reason,
		Reference:    reference,
	}

	// Create movement
	if err := i.DB.Create(movement).Error; err != nil {
		return err
	}

	// Update stock item current stock
	if movementType == "in" || movementType == "adjustment" {
		i.DB.Model(&StockItem{}).Where("id = ?", stockItemID).
			Update("current_stock", gorm.Expr("current_stock + ?", quantity))
	} else if movementType == "out" || movementType == "wastage" {
		i.DB.Model(&StockItem{}).Where("id = ?", stockItemID).
			Update("current_stock", gorm.Expr("current_stock - ?", quantity))
	}

	// Check if low stock
	var stockItem StockItem
	i.DB.First(&stockItem, stockItemID)
	if stockItem.CurrentStock <= stockItem.MinimumStock {
		i.DB.Model(&StockItem{}).Where("id = ?", stockItemID).Update("is_low_stock", true)
	} else {
		i.DB.Model(&StockItem{}).Where("id = ?", stockItemID).Update("is_low_stock", false)
	}

	return nil
}

// CheckLowStockAlerts checks for low stock items
func (i *InventoryService) CheckLowStockAlerts() ([]StockItem, error) {
	var items []StockItem
	if err := i.DB.Where("is_low_stock = ?", true).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
