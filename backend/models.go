package main

import (
	"time"
)

// ========================================
// MODELS
// ========================================

// User model
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"` // Never send password in JSON
	Name      string    `json:"name" gorm:"not null"`
	NameAr    string    `json:"name_ar"`
	Role      string    `json:"role" gorm:"not null;default:'staff'"`
	Phone     string    `json:"phone"`
	Avatar    string    `json:"avatar"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RestaurantSettings model
type RestaurantSettings struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	Name             string    `json:"name" gorm:"not null"`
	NameAr           string    `json:"name_ar" gorm:"not null"`
	Logo             string    `json:"logo"`
	Address          string    `json:"address"`
	Phone            string    `json:"phone"`
	Email            string    `json:"email"`
	Currency         string    `json:"currency" gorm:"default:'EGP'"`
	CurrencySymbol   string    `json:"currency_symbol" gorm:"default:'ج.م'"`
	TaxRate          float64   `json:"tax_rate" gorm:"default:0.14"`
	ServiceCharge    float64   `json:"service_charge" gorm:"default:0.10"`
	Language         string    `json:"language" gorm:"default:'ar'"`
	ThemeColor       string    `json:"theme_color" gorm:"default:'#10b981'"`
	IsOpen           bool      `json:"is_open" gorm:"default:true"`
	ReceiptHeader    string    `json:"receipt_header" gorm:"type:text"`
	ReceiptFooter    string    `json:"receipt_footer" gorm:"type:text"`
	Settings         string    `json:"settings" gorm:"type:json"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Category model
type Category struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	NameAr       string    `json:"name_ar" gorm:"not null"`
	Description   string    `json:"description"`
	DescriptionAr string    `json:"description_ar"`
	Icon         string    `json:"icon"`
	DisplayOrder int       `json:"display_order" gorm:"default:0"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	IsAvailable  bool      `json:"is_available" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	MenuItems    []MenuItem `json:"menu_items,omitempty" gorm:"foreignKey:CategoryID"`
	Modifiers    []Modifier `json:"modifiers,omitempty" gorm:"foreignKey:CategoryID"`
}

// MenuItem model
type MenuItem struct {
	ID              uint    `json:"id" gorm:"primaryKey"`
	Name            string  `json:"name" gorm:"not null"`
	NameAr          string  `json:"name_ar" gorm:"not null"`
	Description     string  `json:"description"`
	DescriptionAr   string  `json:"description_ar"`
	SKU             string  `json:"sku" gorm:"uniqueIndex"`
	Barcode         string  `json:"barcode"`
	CategoryID      uint    `json:"category_id" gorm:"not null"`
	Category        Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Price           float64 `json:"price" gorm:"not null"`
	CostPrice       float64 `json:"cost_price"`
	DiscountPrice   float64 `json:"discount_price"`
	DiscountType    string  `json:"discount_type"`
	DiscountUntil   *time.Time `json:"discount_until"`
	IsAvailable     bool    `json:"is_available" gorm:"default:true"`
	StockQuantity   int     `json:"stock_quantity" gorm:"default:0"`
	LowStockAlert  int     `json:"low_stock_alert" gorm:"default:10"`
	Image           string  `json:"image"`
	Color           string  `json:"color"`
	PreparationTime int     `json:"preparation_time"` // in minutes
	IsModifierOnly  bool    `json:"is_modifier_only" gorm:"default:false"`
	OrderCount      int     `json:"order_count" gorm:"default:0"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Modifier model
type Modifier struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	Name        string          `json:"name" gorm:"not null"`
	NameAr      string          `json:"name_ar" gorm:"not null"`
	Price       float64         `json:"price" gorm:"default:0"`
	Type        string          `json:"type" gorm:"not null;default:'optional'"`
	MinSelect   int             `json:"min_select" gorm:"default:0"`
	MaxSelect   *int            `json:"max_select"`
	CategoryID  *uint           `json:"category_id"`
	Category    *Category       `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	MenuItemID  *uint           `json:"menu_item_id"`
	MenuItem    *MenuItem        `json:"menu_item,omitempty" gorm:"foreignKey:MenuItemID"`
	Options     []ModifierOption `json:"options,omitempty" gorm:"foreignKey:ModifierID"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ModifierOption model
type ModifierOption struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"not null"`
	NameAr     string    `json:"name_ar" gorm:"not null"`
	Price      float64   `json:"price" gorm:"default:0"`
	IsDefault  bool      `json:"is_default" gorm:"default:false"`
	ModifierID uint      `json:"modifier_id" gorm:"not null"`
	Modifier   Modifier  `json:"modifier,omitempty" gorm:"foreignKey:ModifierID"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Combo model
type Combo struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"not null"`
	NameAr        string    `json:"name_ar" gorm:"not null"`
	Description   string    `json:"description"`
	DescriptionAr string    `json:"description_ar"`
	Items         string    `json:"items" gorm:"type:json;not null"` // JSON array
	Price         float64   `json:"price" gorm:"not null"`
	DiscountPrice float64   `json:"discount_price"`
	IsAvailable   bool      `json:"is_available" gorm:"default:true"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Table model
type Table struct {
	ID              uint        `json:"id" gorm:"primaryKey"`
	Number          string      `json:"number" gorm:"uniqueIndex;not null"`
	Name            string      `json:"name"`
	NameAr          string      `json:"name_ar"`
	Section         string      `json:"section"`
	Capacity        int         `json:"capacity" gorm:"default:4"`
	QRCode          string      `json:"qr_code"`
	Status          string      `json:"status" gorm:"not null;default:'available'"`
	CurrentOrderID  *uint       `json:"current_order_id" gorm:"uniqueIndex"`
	CurrentOrder    *Order      `json:"current_order,omitempty" gorm:"foreignKey:CurrentOrderID"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// Reservation model
type Reservation struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	CustomerName    string     `json:"customer_name" gorm:"not null"`
	CustomerPhone   string     `json:"customer_phone" gorm:"not null"`
	TableID        *uint      `json:"table_id"`
	Table          *Table     `json:"table,omitempty" gorm:"foreignKey:TableID"`
	PartySize       int        `json:"party_size" gorm:"not null"`
	ReservationTime time.Time  `json:"reservation_time" gorm:"not null"`
	Status         string     `json:"status" gorm:"not null;default:'pending'"`
	Notes          string     `json:"notes"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Order model
type Order struct {
	ID              uint         `json:"id" gorm:"primaryKey"`
	OrderNumber     string       `json:"order_number" gorm:"uniqueIndex;not null"`
	TableID         *uint        `json:"table_id"`
	Table           *Table       `json:"table,omitempty" gorm:"foreignKey:TableID"`
	UserID          uint         `json:"user_id" gorm:"not null"`
	Type            string       `json:"type" gorm:"not null;default:'dine_in'"`
	Status          string       `json:"status" gorm:"not null;default:'pending'"`
	Priority        string       `json:"priority" gorm:"not null;default:'normal'"`
	CustomerName    string       `json:"customer_name"`
	CustomerPhone   string       `json:"customer_phone"`
	CustomerAddress string       `json:"customer_address"`
	Subtotal        float64      `json:"subtotal" gorm:"default:0"`
	TaxAmount       float64      `json:"tax_amount" gorm:"default:0"`
	ServiceCharge   float64      `json:"service_charge" gorm:"default:0"`
	Discount        float64      `json:"discount" gorm:"default:0"`
	Total           float64      `json:"total" gorm:"default:0"`
	PaidAmount      float64      `json:"paid_amount" gorm:"default:0"`
	Remaining       float64      `json:"remaining" gorm:"default:0"`
	PaymentStatus   string       `json:"payment_status" gorm:"not null;default:'unpaid'"`
	PaymentMethod   string       `json:"payment_method"`
	PaymentReference string       `json:"payment_reference"`
	CreatedAt       time.Time    `json:"created_at"`
	StartedAt       *time.Time   `json:"started_at"`
	CompletedAt     *time.Time   `json:"completed_at"`
	CancelledAt     *time.Time   `json:"cancelled_at"`
	Notes           string       `json:"notes" gorm:"type:text"`
	KitchenNotes    string       `json:"kitchen_notes" gorm:"type:text"`
	VoiceNoteURL    string       `json:"voice_note_url"`
	Items           []OrderItem  `json:"items,omitempty" gorm:"foreignKey:OrderID"`
	Payments        []Payment    `json:"payments,omitempty" gorm:"foreignKey:OrderID"`
}

// OrderItem model
type OrderItem struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	OrderID       uint      `json:"order_id" gorm:"not null"`
	Order         *Order    `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	MenuItemID    uint      `json:"menu_item_id" gorm:"not null"`
	MenuItemName  string    `json:"menu_item_name" gorm:"not null"`
	Quantity      int       `json:"quantity" gorm:"default:1"`
	UnitPrice     float64   `json:"unit_price" gorm:"not null"`
	Modifiers     string    `json:"modifiers" gorm:"type:json"`
	Status        string    `json:"status" gorm:"not null;default:'pending'"`
	Notes         string    `json:"notes" gorm:"type:text"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Payment model
type Payment struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	OrderID         uint      `json:"order_id" gorm:"not null"`
	Order           *Order    `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	UserID          uint      `json:"user_id" gorm:"not null"`
	Type            string    `json:"type" gorm:"not null;default:'sale'"`
	Method          string    `json:"method" gorm:"not null;default:'cash'"`
	Amount          float64   `json:"amount" gorm:"not null"`
	Reference       string    `json:"reference"`
	CashTendered    float64   `json:"cash_tendered"`
	ChangeAmount    float64   `json:"change_amount"`
	CreatedAt       time.Time `json:"created_at"`
}

// StockItem model
type StockItem struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	Name          string     `json:"name" gorm:"not null"`
	NameAr        string     `json:"name_ar" gorm:"not null"`
	SKU           string     `json:"sku" gorm:"uniqueIndex"`
	Unit          string     `json:"unit"`
	CurrentStock  float64    `json:"current_stock" gorm:"default:0"`
	MinimumStock  float64    `json:"minimum_stock" gorm:"default:0"`
	CostPerUnit   float64    `json:"cost_per_unit"`
	Supplier      string     `json:"supplier"`
	IsLowStock    bool       `json:"is_low_stock" gorm:"default:false"`
	LastReorderAt *time.Time `json:"last_reorder_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Movements     []StockMovement `json:"movements,omitempty" gorm:"foreignKey:StockItemID"`
}

// StockMovement model
type StockMovement struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	StockItemID uint       `json:"stock_item_id" gorm:"not null"`
	StockItem   *StockItem `json:"stock_item,omitempty" gorm:"foreignKey:StockItemID"`
	Type        string     `json:"type" gorm:"not null"`
	Quantity    float64    `json:"quantity" gorm:"not null"`
	CostPerUnit float64    `json:"cost_per_unit"`
	Reason      string     `json:"reason" gorm:"type:text"`
	Reference   string     `json:"reference"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Shift model
type Shift struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	StartTime   time.Time `json:"start_time" gorm:"default:CURRENT_TIMESTAMP"`
	EndTime     *time.Time `json:"end_time"`
	OrdersCount int       `json:"orders_count" gorm:"default:0"`
	TotalSales  float64   `json:"total_sales" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
}

// Customer model
type Customer struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	Name          string            `json:"name" gorm:"not null"`
	Phone         string            `json:"phone" gorm:"uniqueIndex;not null"`
	Email         string            `json:"email"`
	Address       string            `json:"address"`
	Points        int               `json:"points" gorm:"default:0"`
	TotalSpent    float64           `json:"total_spent" gorm:"default:0"`
	VisitsCount   int               `json:"visits_count" gorm:"default:0"`
	Birthday      *time.Time        `json:"birthday"`
	Preferences   string            `json:"preferences" gorm:"type:json"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	LoyaltyTransactions []LoyaltyTransaction `json:"loyalty_transactions,omitempty" gorm:"foreignKey:CustomerID"`
}

// LoyaltyTransaction model
type LoyaltyTransaction struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	CustomerID uint      `json:"customer_id" gorm:"not null"`
	Customer   *Customer `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Type       string    `json:"type" gorm:"not null"` // "earned", "redeemed"
	Points     int       `json:"points" gorm:"not null"`
	OrderID    *uint     `json:"order_id"`
	Notes      string    `json:"notes" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at"`
}

// Discount model
type Discount struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	Name         string     `json:"name" gorm:"not null"`
	NameAr       string     `json:"name_ar" gorm:"not null"`
	Type         string     `json:"type" gorm:"not null"` // "percentage", "fixed", "buy_x_get_y"
	Value        float64    `json:"value" gorm:"not null"`
	MinOrderAmount *float64  `json:"min_order_amount"`
	ApplyTo      string     `json:"apply_to" gorm:"not null;default:'all'"` // "all", "categories", "items"`
	ApplyToIDs   string     `json:"apply_to_ids" gorm:"type:json"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	MaxUsage     *int       `json:"max_usage"`
	UsedCount    int        `json:"used_count" gorm:"default:0"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	Code         string     `json:"code" gorm:"uniqueIndex"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ReceiptTemplate model
type ReceiptTemplate struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	IsDefault    bool      `json:"is_default" gorm:"default:false"`
	TemplateHTML string    `json:"template_html" gorm:"type:text;not null"`
	CSSStyles    string    `json:"css_styles" gorm:"type:text"`
	ShowLogo     bool      `json:"show_logo" gorm:"default:true"`
	ShowQR        bool      `json:"show_qr" gorm:"default:false"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Printer model
type Printer struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	Name            string    `json:"name" gorm:"not null"`
	Type            string    `json:"type" gorm:"not null"` // "receipt", "kitchen", "bar"
	IPAddress       string    `json:"ip_address"`
	Port            int       `json:"port"`
	IsDefault       bool      `json:"is_default" gorm:"default:false"`
	IsActive        bool      `json:"is_active" gorm:"default:true"`
	PrintCategories string    `json:"print_categories" gorm:"type:json"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// DailyReport model
type DailyReport struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	ReportDate       time.Time `json:"report_date" gorm:"uniqueIndex;not null"`
	TotalOrders      int       `json:"total_orders" gorm:"default:0"`
	TotalRevenue     float64   `json:"total_revenue" gorm:"default:0"`
	TotalCost        float64   `json:"total_cost" gorm:"default:0"`
	GrossProfit      float64   `json:"gross_profit" gorm:"default:0"`
	DineInOrders    int       `json:"dine_in_orders" gorm:"default:0"`
	TakeawayOrders   int       `json:"takeaway_orders" gorm:"default:0"`
	DeliveryOrders   int       `json:"delivery_orders" gorm:"default:0"`
	CashPayments     float64   `json:"cash_payments" gorm:"default:0"`
	CardPayments     float64   `json:"card_payments" gorm:"default:0"`
	WalletPayments   float64   `json:"wallet_payments" gorm:"default:0"`
	SalesByCategory string    `json:"sales_by_category" gorm:"type:json"`
	TopItems        string    `json:"top_items" gorm:"type:json"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// AuditLog model
type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Action    string    `json:"action" gorm:"not null"`
	Entity    string    `json:"entity"`
	EntityID  *uint     `json:"entity_id"`
	Changes   string    `json:"changes" gorm:"type:json"`
	Metadata  string    `json:"metadata" gorm:"type:json"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
}

// Request/Response DTOs
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateOrderRequest struct {
	TableID         *uint                 `json:"table_id"`
	Type            string                 `json:"type"`
	Priority        string                 `json:"priority"`
	CustomerName    string                 `json:"customer_name"`
	CustomerPhone   string                 `json:"customer_phone"`
	CustomerAddress string                 `json:"customer_address"`
	Items           []CreateOrderItemRequest `json:"items" binding:"required"`
	Notes           string                 `json:"notes"`
	KitchenNotes    string                 `json:"kitchen_notes"`
}

type CreateOrderItemRequest struct {
	MenuItemID uint   `json:"menu_item_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required"`
	Modifiers  string `json:"modifiers"`
	Notes      string `json:"notes"`
}

type PaymentRequest struct {
	OrderID       uint    `json:"order_id" binding:"required"`
	Method        string  `json:"method" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
	CashTendered  float64 `json:"cash_tendered"`
	Reference     string  `json:"reference"`
}

type WhatsAppMessageRequest struct {
	To      string `json:"to" binding:"required"`
	Message string `json:"message" binding:"required"`
}

// Response wrappers
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
