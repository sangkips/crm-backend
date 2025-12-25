# Investify API

Go backend API for the Investify Inventory Management System.

## Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Make (optional, for using Makefile commands)

## Getting Started

1. **Clone the repository and navigate to the API directory:**
   ```bash
   cd investify-api
   ```

2. **Copy the environment file:**
   ```bash
   cp .env.example .env
   ```

3. **Update the `.env` file with your database credentials:**
   ```env
   DB_HOST=localhost
   DB_PORT=5432
   DB_NAME=investify
   DB_USER=postgres
   DB_PASSWORD=your_password
   ```

4. **Download dependencies:**
   ```bash
   go mod download
   go mod tidy
   ```

5. **Run the application:**
   ```bash
   go run cmd/api/main.go
   ```

   Or using Make:
   ```bash
   make run
   ```

6. **The API will be available at:**
   ```
   http://localhost:8080
   ```

## API Endpoints

### Health Check
- `GET /health` - Health check endpoint

### Authentication
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/register` - Register
- `POST /api/v1/auth/refresh` - Refresh token
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password

### Products (requires `manage-products` permission)
- `GET /api/v1/products` - List products
- `POST /api/v1/products` - Create product
- `GET /api/v1/products/:slug` - Get product
- `PUT /api/v1/products/:slug` - Update product
- `DELETE /api/v1/products/:slug` - Delete product

### Orders (requires `manage-orders` permission)
- `GET /api/v1/orders` - List orders
- `POST /api/v1/orders` - Create order
- `GET /api/v1/orders/:id` - Get order
- `PUT /api/v1/orders/:id` - Update order
- `DELETE /api/v1/orders/:id/cancel` - Cancel order

### Purchases (requires `manage-purchases` permission)
- `GET /api/v1/purchases` - List purchases
- `POST /api/v1/purchases` - Create purchase
- `GET /api/v1/purchases/:id` - Get purchase
- `PUT /api/v1/purchases/:id` - Update purchase
- `DELETE /api/v1/purchases/:id` - Delete purchase
- `POST /api/v1/purchases/:id/approve` - Approve purchase

### Quotations (requires `manage-quotations` permission)
- `GET /api/v1/quotations` - List quotations
- `POST /api/v1/quotations` - Create quotation
- `GET /api/v1/quotations/:id` - Get quotation
- `PUT /api/v1/quotations/:id` - Update quotation
- `DELETE /api/v1/quotations/:id` - Delete quotation

### Customers (requires `manage-customers` permission)
- `GET /api/v1/customers` - List customers
- `POST /api/v1/customers` - Create customer
- `GET /api/v1/customers/:id` - Get customer
- `PUT /api/v1/customers/:id` - Update customer
- `DELETE /api/v1/customers/:id` - Delete customer

### Suppliers (requires `manage-suppliers` permission)
- `GET /api/v1/suppliers` - List suppliers
- `POST /api/v1/suppliers` - Create supplier
- `GET /api/v1/suppliers/:id` - Get supplier
- `PUT /api/v1/suppliers/:id` - Update supplier
- `DELETE /api/v1/suppliers/:id` - Delete supplier

### Categories (requires `manage-categories` permission)
- `GET /api/v1/categories` - List categories
- `POST /api/v1/categories` - Create category
- `PUT /api/v1/categories/:id` - Update category
- `DELETE /api/v1/categories/:id` - Delete category

### Units (requires `manage-units` permission)
- `GET /api/v1/units` - List units
- `POST /api/v1/units` - Create unit
- `PUT /api/v1/units/:id` - Update unit
- `DELETE /api/v1/units/:id` - Delete unit

### Profile
- `GET /api/v1/profile` - Get current user profile
- `PUT /api/v1/profile` - Update profile
- `PUT /api/v1/profile/password` - Change password
- `GET /api/v1/profile/settings` - Get settings
- `PUT /api/v1/profile/store-settings` - Update store settings

### Admin (requires `admin` or `super-admin` role)
- `GET /api/v1/admin/users` - List users
- `POST /api/v1/admin/users` - Create user
- `GET /api/v1/admin/users/:id` - Get user
- `PUT /api/v1/admin/users/:id` - Update user
- `DELETE /api/v1/admin/users/:id` - Delete user
- `GET /api/v1/admin/roles` - List roles
- `POST /api/v1/admin/roles` - Create role
- `PUT /api/v1/admin/roles/:id` - Update role
- `PUT /api/v1/admin/roles/:id/permissions` - Update role permissions
- `GET /api/v1/admin/permissions` - List permissions

### Reports (requires `view-reports` permission)
- `GET /api/v1/reports/orders` - Orders report
- `POST /api/v1/reports/orders/export` - Export orders report
- `GET /api/v1/reports/purchases` - Purchases report
- `POST /api/v1/reports/purchases/export` - Export purchases report

## Project Structure

```
investify-api/
├── cmd/api/                 # Application entry point
├── internal/
│   ├── config/              # Configuration
│   ├── domain/
│   │   ├── entity/          # Domain entities
│   │   ├── enum/            # Enums
│   │   └── repository/      # Repository interfaces
│   ├── application/
│   │   └── service/         # Business logic services
│   ├── infrastructure/
│   │   ├── database/        # Database connection
│   │   └── repository/      # Repository implementations
│   └── presentation/
│       └── http/
│           ├── handler/     # HTTP handlers
│           ├── middleware/  # HTTP middleware
│           └── dto/         # Request/Response DTOs
├── pkg/
│   ├── apperror/            # Application errors
│   ├── pagination/          # Pagination utilities
│   └── utils/               # Utility functions
├── migrations/              # Database migrations
└── docs/                    # Documentation
```

## Development

### Running with hot reload
```bash
# Install air first
go install github.com/air-verse/air@latest

# Run with hot reload
make dev
```

### Running tests
```bash
make test
```

### Building
```bash
make build
```

## License

MIT
