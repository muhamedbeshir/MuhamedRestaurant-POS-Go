// Muhamed Restaurant POS - Frontend Application
// Complete SPA with all pages and features

class RestaurantPOS {
    constructor() {
        this.state = {
            currentPage: 'dashboard',
            currentLang: 'ar',
            settings: {},
            categories: [],
            items: [],
            tables: [],
            orders: [],
            cart: [],
            selectedCategory: null,
            selectedTable: null,
            stats: {},
            user: null,
            token: localStorage.getItem('token') || null
        };

        this.translations = {};
        this.init();
    }

    async init() {
        console.log('🚀 Restaurant POS initializing...');

        // Load settings
        await this.loadSettings();

        // Load translations
        this.translations = await this.getTranslations();

        // Load data
        await this.loadDashboardStats();

        // Setup UI
        this.setupNavigation();
        this.render();
        this.updateLanguage();

        console.log('✅ Restaurant POS ready!');
    }

    // ========================================
    // API CALLS
    // ========================================

    async apiRequest(endpoint, method = 'GET', body = null) {
        const options = {
            method: method,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${this.state.token}`
            }
        };

        if (body) {
            options.body = JSON.stringify(body);
        }

        try {
            const response = await fetch(`/api${endpoint}`, options);
            return await response.json();
        } catch (error) {
            console.error('API Error:', error);
            return { success: false, error: error.message };
        }
    }

    async loadSettings() {
        const result = await this.apiRequest('/settings');
        if (result.success) {
            this.state.settings = result.data;
            this.state.currentLang = result.data.language || 'ar';
            document.documentElement.style.setProperty('--theme-color', result.data.theme_color || '#10b981');
        }
    }

    async getTranslations(lang = this.state.currentLang) {
        return {
            ar: {
                dashboard: 'لوحة التحكم',
                pos: 'شبكة البيع',
                tables: 'الطاولات',
                kitchen: 'المطبخ',
                orders: 'الطلبات',
                payments: 'المدفوعات',
                reports: 'التقارير',
                inventory: 'المخزون',
                staff: 'الموظفين',
                customers: 'العملاء',
                reservations: 'الحجوزات',
                discounts: 'الخصومات',
                settings: 'الإعدادات',
                total_orders: 'إجمالي الطلبات اليوم',
                total_revenue: 'إجمالي المبيعات اليوم',
                available_tables: 'الطاولات المتاحة',
                low_stock: 'المخزون المنخفض',
                loading: 'جاري التحميل...'
            },
            en: {
                dashboard: 'Dashboard',
                pos: 'Point of Sale',
                tables: 'Tables',
                kitchen: 'Kitchen',
                orders: 'Orders',
                payments: 'Payments',
                reports: 'Reports',
                inventory: 'Inventory',
                staff: 'Staff',
                customers: 'Customers',
                reservations: 'Reservations',
                discounts: 'Discounts',
                settings: 'Settings',
                total_orders: 'Total Orders Today',
                total_revenue: 'Total Revenue Today',
                available_tables: 'Available Tables',
                low_stock: 'Low Stock',
                loading: 'Loading...'
            }
        }[lang] || this.translations.ar;
    }

    async loadDashboardStats() {
        const result = await this.apiRequest('/dashboard/stats');
        if (result.success) {
            this.state.stats = result.data;
            this.updateDashboardStats();
        }
    }

    // ========================================
    // NAVIGATION
    // ========================================

    setupNavigation() {
        document.querySelectorAll('.nav-item').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const page = item.dataset.page;
                this.navigate(page);
            });
        });
    }

    navigate(page) {
        this.state.currentPage = page;

        // Update nav items
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.toggle('active', item.dataset.page === page);
        });

        // Show page
        document.querySelectorAll('.page').forEach(pageEl => {
            pageEl.classList.toggle('active', pageEl.id === `page-${page}`);
        });

        // Update title
        const title = this.translations[this.state.currentLang][page] || page;
        document.querySelector('#page-title').textContent = title;

        // Load page data
        this.loadPageData(page);
    }

    async loadPageData(page) {
        switch (page) {
            case 'dashboard':
                await this.loadDashboardStats();
                break;
            case 'pos':
                await this.loadPOSData();
                break;
            case 'tables':
                await this.loadTablesData();
                break;
            case 'kitchen':
                await this.loadKitchenData();
                break;
            case 'orders':
                await this.loadOrdersData();
                break;
        }
    }

    // ========================================
    // PAGE LOADERS
    // ========================================

    async loadPOSData() {
        // Load categories
        const catsResult = await this.apiRequest('/menu/categories');
        if (catsResult.success) {
            this.state.categories = catsResult.data || [];
            this.renderCategories();
        }

        // Load items
        const itemsResult = await this.apiRequest('/menu/items');
        if (itemsResult.success) {
            this.state.items = itemsResult.data || [];
            this.renderItems();
        }
    }

    async loadTablesData() {
        const result = await this.apiRequest('/tables');
        if (result.success) {
            this.state.tables = result.data || [];
            this.renderTables();
        }
    }

    async loadKitchenData() {
        // Load pending orders
        const result = await this.apiRequest('/orders?status=preparing');
        if (result.success) {
            this.state.orders = result.data || [];
            this.renderKitchenOrders();
        }
    }

    async loadOrdersData() {
        const result = await this.apiRequest('/orders');
        if (result.success) {
            this.state.orders = result.data || [];
            this.renderOrders();
        }
    }

    // ========================================
    // RENDERING
    // ========================================

    render() {
        this.updateDashboardStats();
    }

    updateDashboardStats() {
        document.querySelector('#stat-orders').textContent = this.state.stats.total_orders || 0;
        document.querySelector('#stat-revenue').textContent = (this.state.stats.total_revenue || 0) + ' ج.م';
        document.querySelector('#stat-tables').textContent = this.state.stats.available_tables || 0;
        document.querySelector('#stock-alerts').textContent = this.state.stats.low_stock || 0;
    }

    renderCategories() {
        const container = document.querySelector('.categories');
        if (!container) return;

        container.innerHTML = `
            <button class="category-btn active" data-category="all">${this.translations[this.state.currentLang].all || 'الكل'}</button>
        ` + this.state.categories.map(cat => `
            <button class="category-btn" data-category="${cat.id}">${cat.name_ar || cat.name}</button>
        `).join('');

        // Add event listeners
        container.querySelectorAll('.category-btn').forEach(btn => {
            btn.addEventListener('click', () => this.selectCategory(btn.dataset.category));
        });
    }

    renderItems() {
        const container = document.querySelector('.items-grid');
        if (!container) return;

        const filteredItems = this.state.selectedCategory === 'all' || !this.state.selectedCategory
            ? this.state.items
            : this.state.items.filter(item => item.category_id == this.state.selectedCategory);

        container.innerHTML = filteredItems.map(item => `
            <div class="item-card" data-item-id="${item.id}">
                <div class="item-image">${item.image ? `<img src="${item.image}">` : '🍽️'}</div>
                <div class="item-info">
                    <div class="item-name">${item.name_ar || item.name}</div>
                    <div class="item-price">${item.price.toFixed(2)} ج.م</div>
                </div>
            </div>
        `).join('');
    }

    renderTables() {
        const container = document.querySelector('.tables-grid');
        if (!container) return;

        container.innerHTML = this.state.tables.map(table => `
            <div class="table-card ${table.status}" data-table-id="${table.id}">
                <div class="table-number">${table.number}</div>
                <div class="table-capacity">${table.capacity} أشخاص</div>
            </div>
        `).join('');
    }

    renderKitchenOrders() {
        const container = document.querySelector('.kitchen-orders');
        if (!container) return;

        container.innerHTML = this.state.orders.map(order => `
            <div class="kitchen-order-card ${order.priority}">
                <div class="order-number">#${order.order_number}</div>
                <div class="order-time">${new Date(order.created_at).toLocaleTimeString()}</div>
                <div class="order-items">${order.items ? order.items.map(i => i.menu_item_name).join(', ') : ''}</div>
            </div>
        `).join('');
    }

    renderOrders() {
        const container = document.querySelector('.orders-list');
        if (!container) return;

        container.innerHTML = this.state.orders.map(order => `
            <div class="order-card" data-order-id="${order.id}">
                <div class="order-header">
                    <span class="order-number">#${order.order_number}</span>
                    <span class="order-status ${order.status}">${order.status}</span>
                </div>
                <div class="order-total">${order.total.toFixed(2)} ج.م</div>
                <div class="order-time">${new Date(order.created_at).toLocaleTimeString()}</div>
            </div>
        `).join('');
    }

    renderCart() {
        const container = document.querySelector('.cart-items');
        if (!container) return;

        if (this.state.cart.length === 0) {
            container.innerHTML = `<div class="empty-cart">السلة فارغة</div>`;
            return;
        }

        container.innerHTML = this.state.cart.map((item, index) => `
            <div class="cart-item" data-cart-index="${index}">
                <div class="item-info">
                    <div class="item-name">${item.name_ar || item.name}</div>
                    <div class="item-qty">${item.quantity} × ${item.price.toFixed(2)}</div>
                </div>
                <div class="item-controls">
                    <button class="qty-btn minus">-</button>
                    <button class="qty-btn plus">+</button>
                </div>
            </div>
        `).join('');

        this.updateCartTotals();
    }

    updateCartTotals() {
        const subtotal = this.state.cart.reduce((sum, item) => sum + (item.price * item.quantity), 0);
        const tax = subtotal * (this.state.settings.tax_rate || 0.14);
        const total = subtotal + tax;

        document.querySelector('#cart-subtotal').textContent = subtotal.toFixed(2) + ' ج.م';
        document.querySelector('#cart-tax').textContent = tax.toFixed(2) + ' ج.م';
        document.querySelector('#cart-total').textContent = total.toFixed(2) + ' ج.م';
    }

    // ========================================
    // ACTIONS
    // ========================================

    selectCategory(categoryId) {
        this.state.selectedCategory = categoryId;
        this.renderCategories();
        this.renderItems();
    }

    addToCart(itemId) {
        const item = this.state.items.find(i => i.id === itemId);
        if (!item) return;

        const existing = this.state.cart.find(c => c.id === itemId);
        if (existing) {
            existing.quantity++;
        } else {
            this.state.cart.push({
                id: item.id,
                name: item.name,
                name_ar: item.name_ar,
                price: item.price,
                quantity: 1
            });
        }

        this.renderCart();
    }

    updateCartQuantity(index, delta) {
        const item = this.state.cart[index];
        if (!item) return;

        item.quantity += delta;

        if (item.quantity <= 0) {
            this.state.cart.splice(index, 1);
        }

        this.renderCart();
    }

    async checkout() {
        if (this.state.cart.length === 0) {
            alert('السلة فارغة!');
            return;
        }

        const order = {
            items: this.state.cart.map(item => ({
                menu_item_id: item.id,
                quantity: item.quantity
            })),
            notes: ''
        };

        const result = await this.apiRequest('/orders', 'POST', order);

        if (result.success) {
            this.state.cart = [];
            this.renderCart();
            alert(`تم إنشاء الطلب #${result.data.order_number} بنجاح!`);
        }
    }

    // ========================================
    // LANGUAGE
    // ========================================

    updateLanguage() {
        const lang = this.state.currentLang;

        // Update direction
        document.documentElement.lang = lang;
        document.documentElement.dir = lang === 'ar' ? 'rtl' : 'ltr';

        // Update UI
        document.querySelectorAll('[data-key]').forEach(el => {
            const key = el.dataset.key;
            if (this.translations[lang] && this.translations[lang][key]) {
                el.textContent = this.translations[lang][key];
            }
        });
    }

    async setLanguage(lang) {
        this.state.currentLang = lang;
        this.state.settings.language = lang;

        await this.apiRequest('/settings', 'PUT', this.state.settings);

        this.translations = await this.getTranslations(lang);
        this.updateLanguage();
    }
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.posApp = new RestaurantPOS();
});
