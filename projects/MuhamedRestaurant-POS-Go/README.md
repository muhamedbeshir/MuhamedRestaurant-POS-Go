# Muhamed Restaurant POS - Complete Modular System

## 🎯 Overview

Full-featured Restaurant POS System built with **Modular Architecture**
- 🏗️ **Modular Design** - Backend, Frontend, Database all separated
- 🗄️ **MySQL Database** - Production-ready database
- 🌐 **REST API** - Complete API endpoints
- 💻 **Frontend** - Separated SPA (Single Page Application)
- 🖥️ **Desktop** - Wails wrapper for EXE distribution
- 🌐 **Bilingual** - Arabic + English
- 📦 **Offline Ready** - Can work offline with API caching

---

## 📦 Architecture

```
MuhamedRestaurant-POS-Go/
├── backend/              # Go Backend (REST API)
│   ├── main.go
│   ├── config/
│   ├── models/
│   ├── handlers/
│   ├── services/
│   ├── database/
│   ├── middleware/
│   └── utils/
│
├── frontend/             # Frontend (Separated)
│   ├── index.html
│   ├── css/
│   ├── js/
│   ├── assets/
│   └── pages/
│
├── desktop/              # Wails Desktop Wrapper
│   ├── main.go
│   └── app.go
│
├── database/             # MySQL Scripts
│   ├── schema.sql
│   ├── seeds.sql
│   └── migrations/
│
├── docs/                 # Documentation
│   ├── API.md
│   ├── DEPLOYMENT.md
│   └── DATABASE.md
│
├── deploy/               # Deployment
│   ├── docker/
│   ├── scripts/
│   └── systemd/
│
└── README.md
```

---

## 🚀 Features

### Phase 1: Core (✅ Complete)
- ✅ Authentication & Authorization
- ✅ Settings Management
- ✅ Menu Management (Categories, Items)
- ✅ Orders & Order Items
- ✅ Tables Management
- ✅ Basic Reports
- ✅ Bilingual (Ar/En)

### Phase 2: Enhanced (✅ Complete)
- ✅ Kitchen Display System (KDS)
- ✅ Payments Processing
- ✅ Modifiers & Combos
- ✅ Advanced Reports
- ✅ Inventory Basics
- ✅ Printer Support

### Phase 3: Advanced (✅ Complete)
- ✅ Full Inventory Management
- ✅ Staff Management & Shifts
- ✅ Customer Database & Loyalty
- ✅ Reservations System
- ✅ Discounts & Promotions
- ✅ Receipt Templates

### Phase 4: Enterprise (✅ Complete)
- ✅ Multi-location Support
- ✅ Advanced Integrations
- ✅ Cloud Sync (Optional)
- ✅ Mobile Companion App
- ✅ Advanced Analytics

---

## 🗄️ Database (MySQL)

### Tables (18+ tables)
- users, roles, permissions
- restaurant_settings
- categories, menu_items, modifiers, combos
- tables, reservations
- orders, order_items
- payments, transactions
- inventory, stock_movements
- staff, shifts
- customers, loyalty_points
- discounts, promotions
- receipt_templates, printers
- daily_reports, audit_logs

---

## 📱 Frontend

### Pages (15+ pages)
1. POS (Point of Sale)
2. Tables
3. Kitchen Display System (KDS)
4. Orders Management
5. Payments
6. Reports & Analytics
7. Inventory
8. Staff Management
9. Customers
10. Reservations
11. Discounts & Promotions
12. Settings
13. Receipt Templates
14. Printer Configuration
15. Dashboard

### Technologies
- Vanilla JavaScript (ES6+)
- Alpine.js (~15KB)
- TailwindCSS (CDN or embedded)
- Socket.io Client (WebSocket)
- Chart.js (Reports)
- No build step required

---

## 🖥️ Backend (Go)

### Technologies
- Go 1.22+
- Gin Framework (HTTP Router)
- GORM (ORM for MySQL)
- WebSocket (Gorilla WebSocket)
- JWT Authentication
- MySQL Driver

### API Endpoints (100+ endpoints)
- Authentication (Login, Logout, Refresh)
- Settings (CRUD)
- Menu (Categories, Items, Modifiers, Combos)
- Orders (CRUD, Status Update)
- Tables (CRUD, Status)
- Payments (Create, Refund)
- Reports (Daily, Weekly, Monthly, Analytics)
- Inventory (Items, Movements, Alerts)
- Staff (CRUD, Shifts)
- Customers (CRUD, Loyalty)
- Reservations (CRUD)
- Discounts (CRUD, Activate/Deactivate)
- Printers (CRUD, Test Print)

---

## 🚀 Getting Started

### Prerequisites
- Go 1.22+
- MySQL 8.0+
- Node.js (optional - for some tools)

### Setup

#### 1. Clone
```bash
git clone https://github.com/muhamedbeshir/MuhamedRestaurant-POS-Go.git
cd MuhamedRestaurant-POS-Go
```

#### 2. Database Setup
```bash
# Create database
mysql -u root -p
CREATE DATABASE restaurant_pos CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# Import schema
mysql -u root -p restaurant_pos < database/schema.sql

# Import seeds (optional)
mysql -u root -p restaurant_pos < database/seeds.sql
```

#### 3. Backend Setup
```bash
cd backend

# Install dependencies
go mod download

# Copy config
cp config.example.json config.json

# Edit config
nano config.json

# Run
go run main.go

# Or build
go build -o server
./server
```

#### 4. Frontend Setup
```bash
cd frontend

# Open index.html in browser
# Or serve with a simple server
python3 -m http.server 8080
```

#### 5. Desktop (Optional - for EXE)
```bash
cd desktop

# Install Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Build
wails build

# Run
wails dev
```

---

## 🔧 Configuration

### Backend Config (config.json)
```json
{
  "database": {
    "host": "localhost",
    "port": 3306,
    "user": "root",
    "password": "",
    "name": "restaurant_pos"
  },
  "server": {
    "port": 8080,
    "host": "0.0.0.0"
  },
  "jwt": {
    "secret": "your-secret-key",
    "expiration": "24h"
  },
  "printers": {
    "default": "receipt",
    "kitchen": "kitchen"
  }
}
```

---

## 🖨️ Printing

### Supported Printers
- Thermal Printers (EPSON, Star, Custom)
- Network Printers
- USB Printers

### Features
- ✅ Arabic Text Support
- ✅ English Text Support
- ✅ Direct Printing (No Dialog)
- ✅ Custom Receipt Templates
- ✅ Kitchen Printing
- ✅ Bar Printing

---

## 📊 Reports

### Available Reports
- Daily Sales Report
- Weekly/Monthly Report
- Top Selling Items
- Least Selling Items
- Revenue by Category
- Revenue by Payment Method
- Revenue by Staff
- Revenue by Hour
- Low Stock Alert
- Profit Margins

---

## 🌐 API Documentation

Full API documentation in `docs/API.md`

### Example Endpoints

#### Authentication
```bash
POST /api/auth/login
POST /api/auth/logout
GET  /api/auth/me
```

#### Menu
```bash
GET    /api/menu/categories
GET    /api/menu/items
POST   /api/menu/items
PUT    /api/menu/items/:id
DELETE /api/menu/items/:id
```

#### Orders
```bash
GET    /api/orders
POST   /api/orders
PUT    /api/orders/:id
DELETE /api/orders/:id
POST   /api/orders/:id/status
```

#### Payments
```bash
POST /api/payments
POST /api/payments/refund
```

#### Reports
```bash
GET /api/reports/daily
GET /api/reports/weekly
GET /api/reports/monthly
GET /api/reports/items
```

---

## 🔒 Security

- ✅ JWT Authentication
- ✅ Password Hashing (bcrypt)
- ✅ Role-Based Access Control (RBAC)
- ✅ Input Validation
- ✅ SQL Injection Prevention
- ✅ XSS Prevention
- ✅ CORS Configuration
- ✅ Rate Limiting
- ✅ Audit Logging

---

## 📈 Performance

- ✅ Database Indexing
- ✅ Caching (Redis optional)
- ✅ Connection Pooling
- ✅ Efficient Queries
- ✅ Lazy Loading
- ✅ Pagination

---

## 🔄 WebSocket Events

```javascript
// Order events
order.created
order.updated
order.cancelled
order.item_status_changed

// Table events
table.status_changed
table.assigned

// Kitchen events
kitchen.new_order
kitchen.item_ready
kitchen.order_completed

// Payment events
payment.completed
payment.refunded

// Inventory events
stock.low_alert
stock.movement
```

---

## 🐳 Docker Deployment

```bash
# Build
docker-compose build

# Run
docker-compose up -d

# View logs
docker-compose logs -f
```

---

## 📝 Development

### Backend
```bash
cd backend
go run main.go
```

### Frontend
```bash
cd frontend
python3 -m http.server 8080
```

### Desktop
```bash
cd desktop
wails dev
```

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

---

## 📄 License

Copyright 2024 - Muhamed

---

## 👨‍💻 Development

Built with ❤️ by Muhamed

**Ready for Production!** 🚀
