# Investify API — Project Guide

## Overview

Investify API is a **multi-tenant inventory management / CRM backend** built in Go. It provides RESTful APIs for managing products, orders, purchases, quotations, customers, suppliers, and user administration with role-based access control.

## Tech Stack

| Layer          | Technology                                      |
|----------------|--------------------------------------------------|
| Language       | Go 1.24                                          |
| HTTP Framework | Gin (`github.com/gin-gonic/gin`)                 |
| ORM            | GORM (`gorm.io/gorm`) with PostgreSQL driver     |
| Database       | PostgreSQL 14+                                   |
| Auth           | JWT (`golang-jwt/jwt/v5`) + Google OAuth 2.0     |
| Config         | Viper (`spf13/viper`) — reads `.env` files       |
| Email          | SMTP via `net/smtp`                              |
| Containerization | Docker (multi-stage Alpine build)              |

## Architecture

The project follows **Clean Architecture** (Hexagonal / Ports & Adapters):

```
cmd/api/main.go              → Application entry point & DI wiring
internal/
├── config/                  → Configuration loading (Viper/.env)
├── domain/
│   ├── entity/              → Domain models (structs with GORM tags)
│   ├── enum/                → Status enums (order, purchase, quotation, supplier, tax)
│   └── repository/          → Repository interfaces (ports)
├── application/
│   └── service/             → Business logic services
├── infrastructure/
│   ├── database/            → PostgreSQL connection & auto-migration
│   └── repository/          → GORM repository implementations (adapters)
└── presentation/
    └── http/
        ├── handler/         → Gin HTTP handlers
        ├── middleware/      → Auth, CORS, rate limiting, tenant, idempotency, logger
        └── dto/             → Request/Response DTOs
pkg/
├── apperror/                → Custom application error types
├── email/                   → Email service
├── oauth/                   → Google OAuth service
├── pagination/              → Pagination utilities
└── utils/                   → JWT manager, password hashing, UUID helpers
migrations/                  → SQL migration files
```

## Key Domain Entities

`User`, `Tenant`, `Product`, `Order` (with OrderDetails), `Purchase` (with PurchaseDetails), `Quotation` (with QuotationDetails), `Customer`, `Supplier`, `PasswordResetToken`, `Idempotency`, `UserSettings`

## API Route Structure

All routes are under `/api/v1`. Public routes: `/auth/*` (login, register, refresh, forgot/reset password, Google OAuth). Protected routes require JWT via `Authorization: Bearer <token>` header.

**Resource groups** (each guarded by a permission middleware):
- `/products` — CRUD + import + low-stock (`manage-products`)
- `/orders` — CRUD + status + cancel + pay-due (`manage-orders`, uses idempotency middleware on create)
- `/purchases` — CRUD + approve + pending (`manage-purchases`)
- `/quotations` — CRUD (`manage-quotations`)
- `/customers` — CRUD (`manage-customers`)
- `/suppliers` — CRUD (`manage-suppliers`)
- `/categories` — CRUD (`manage-categories`)
- `/units` — CRUD (`manage-units`)
- `/reports` — Orders/purchases/products reports (`view-reports`) — *placeholder*
- `/users`, `/roles`, `/permissions` — Admin user management (`manage-users`)
- `/tenants` — Tenant management (list, create, members, invite)
- `/admin` — Super-admin routes (tenant user assignment, `super-admin` role required)
- `/profile` — Current user profile & settings
- `/dashboard` — Analytics stats
- `/settings` — App settings

## Middleware Stack

1. **Recovery** — Gin panic recovery
2. **Logger** — Request logging
3. **CORS** — Configurable CORS
4. **Auth** — JWT token validation (on protected routes)
5. **Rate Limiter** — Per-tenant token bucket rate limiting
6. **Permission/Role** — `RequirePermission("...")` / `RequireRole("...")` per route group
7. **Idempotency** — Prevents duplicate order creation via `Idempotency-Key` header
8. **Tenant** — Multi-tenancy scoping

## Multi-Tenancy

The app supports multi-tenancy. Users belong to tenants; data is scoped by `tenant_id`. Repository queries use GORM scopes to filter by tenant. A `super-admin` role can assign users across tenants.

## Development Commands

```bash
make run              # go run cmd/api/main.go
make dev              # Hot reload with Air
make build            # Build binary to bin/
make test             # go test -v ./...
make test-coverage    # Tests with coverage report
make deps             # go mod download && tidy
make docker-build     # Docker build
make docker-run       # Docker run with .env
make migrate-up       # Run DB migrations up
make migrate-down     # Roll back DB migrations
make swagger          # Generate Swagger docs
```

## Configuration

All configuration is via environment variables (or `.env` file). See `.env.example` for all available keys. Key sections: App, Database, JWT, Storage, CORS, Rate Limiting, Email (SMTP), Google OAuth.

## Coding Conventions

- **Repository pattern**: Interfaces in `internal/domain/repository/`, implementations in `internal/infrastructure/repository/`
- **Service layer**: All business logic lives in `internal/application/service/` — services receive repositories via constructor injection
- **Handlers**: Thin HTTP handlers in `internal/presentation/http/handler/` — parse request, call service, return response
- **DTOs**: Request/response structs in `internal/presentation/http/dto/`
- **Error handling**: Use `pkg/apperror` for typed application errors; handlers return consistent JSON error responses
- **Naming**: Go standard — camelCase for unexported, PascalCase for exported
- **Slugs**: Products are identified by slug in URLs, other resources by UUID
- **Database migrations**: GORM auto-migrate is used for schema sync; SQL migrations in `migrations/` for complex changes
- **No test files currently exist** — the project uses `make test` with `go test -v ./...`

## Important Patterns

1. **Dependency Injection**: All wiring happens in `cmd/api/main.go` — repositories → services → handlers
2. **GORM Scopes**: Tenant filtering and soft-delete scoping via `internal/infrastructure/repository/scopes.go`
3. **Idempotency**: Order creation supports idempotency keys to prevent duplicate transactions
4. **Password Reset Flow**: Token-based flow with email sending via SMTP
5. **Google OAuth**: Full OAuth 2.0 flow with callback redirect to frontend
6. **Excel Import**: Product bulk import from Excel files (`excelize` library)
