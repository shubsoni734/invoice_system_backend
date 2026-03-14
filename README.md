# Invoice Backend

Multi-tenant SaaS invoice management platform built with Go 1.22+.

## Project Overview

This backend serves as a platform where the **platform owner (SuperAdmin)** manages multiple businesses (organisations), and each business manages their own customers, services, invoices, payments, and templates.

## Tech Stack

- **Language**: Go 1.22+
- **Framework**: Gin
- **Database**: PostgreSQL 16 (Neon)
- **Query Builder**: sqlc
- **Migrations**: Goose
- **Config**: Viper
- **Logging**: Zap (structured JSON)
- **Validation**: go-playground/validator
- **JWT**: golang-jwt/jwt (RS256 only)
- **Password**: bcrypt (cost 12)

## Prerequisites

- Go 1.22+
- PostgreSQL 16
- OpenSSL (for key generation)

## Install Required CLI Tools (One Time Only)

```bash
# Install goose migration tool
go install github.com/pressly/goose/v3/cmd/goose@latest

# Install sqlc query generator
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Install linter (optional)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Generate JWT Signing Keys (One Time Only)

```bash
mkdir keys

# Organisation user JWT keys
openssl genrsa -out keys/org_private.pem 2048
openssl rsa -in keys/org_private.pem -pubout -out keys/org_public.pem

# SuperAdmin JWT keys
openssl genrsa -out keys/sa_private.pem 2048
openssl rsa -in keys/sa_private.pem -pubout -out keys/sa_public.pem
```

## First-Time Setup

```bash
# 1. Clone repository
git clone <repo-url>
cd invoice-backend

# 2. Copy environment file
cp .env.example .env
# Edit .env with your Neon database URL

# 3. Download Go dependencies
go mod download

# 4. Generate JWT keys (see above)

# 5. Run database migrations
# Option A: Using the migration script (Recommended)
.\scripts\run_migrations.ps1        # Windows
./scripts/run_migrations.sh         # Linux/Mac

# Option B: Manual migration
goose -dir internal/db/migrations postgres "$DATABASE_URL" up

# 6. Create SuperAdmin
.\scripts\create_superadmin.ps1     # Windows
./scripts/create_superadmin.sh      # Linux/Mac

# 7. Generate type-safe query code (if using sqlc)
sqlc generate

# 8. Start the server
go run cmd/server/main.go
```

## Quick Start (For Neon Database)

```powershell
# 1. Setup environment
cp .env.example .env
# Add your Neon DATABASE_URL to .env

# 2. Generate JWT keys
go run generate_keys.go

# 3. Run migrations (creates all 19 tables)
.\scripts\run_migrations.ps1

# 4. Create SuperAdmin
.\scripts\create_superadmin.ps1
# Default: superadmin@invoicepro.com / SuperAdmin@123

# 5. Start server
go run cmd/server/main.go
# Server runs on http://localhost:8080

# 6. Test API
curl http://localhost:8080/health
```

**📚 Detailed Guide:** See `DATABASE_SETUP.md` for complete database setup instructions.

## Day-to-Day Commands

### Run the server
```bash
go run cmd/server/main.go
```

### Build a binary
```bash
go build -o bin/server cmd/server/main.go
```

### Run the built binary
```bash
./bin/server
```

### Run all tests
```bash
go test ./... -v -count=1
```

### Run tests with coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Database Migrations

```bash
# Quick: Use the migration script (Recommended)
.\scripts\run_migrations.ps1        # Windows
./scripts/run_migrations.sh         # Linux/Mac

# Manual: Apply all pending migrations
goose -dir internal/db/migrations postgres "$DATABASE_URL" up

# Roll back last migration
goose -dir internal/db/migrations postgres "$DATABASE_URL" down

# Check migration status
goose -dir internal/db/migrations postgres "$DATABASE_URL" status

# Reset database (⚠️ DESTRUCTIVE - drops all tables)
goose -dir internal/db/migrations postgres "$DATABASE_URL" reset
goose -dir internal/db/migrations postgres "$DATABASE_URL" up

# Create new migration
goose -dir internal/db/migrations create migration_name sql
```

**📚 Full Guide:** See `DATABASE_SETUP.md` for troubleshooting and advanced usage.

### Regenerate sqlc code (after query changes)
```bash
sqlc generate
```

### Run linter
```bash
golangci-lint run ./...
```

## API Endpoints

Server runs on `http://localhost:8080`

### System Routes (No Auth)
- `GET /health` - Liveness probe
- `GET /ready` - Readiness probe (checks DB)
- `GET /metrics` - Prometheus metrics

### Organisation User Routes (Base: `/api/v1`)

**Authentication**
- `POST /auth/login` - Login (returns access + refresh tokens)
- `POST /auth/refresh` - Refresh access token
- `POST /auth/logout` - Revoke refresh token
- `GET /auth/me` - Get current user profile
- `PUT /auth/me/password` - Change own password

**Customers**
- `GET /customers` - List (paginated, searchable)
- `POST /customers` - Create
- `GET /customers/:id` - Get by ID
- `PUT /customers/:id` - Update
- `DELETE /customers/:id` - Soft delete

**Services**
- `GET /services` - List (paginated)
- `POST /services` - Create
- `GET /services/:id` - Get by ID
- `PUT /services/:id` - Update
- `DELETE /services/:id` - Soft delete

**Invoice Sessions**
- `GET /invoice-sessions` - List sessions
- `POST /invoice-sessions` - Create session for year
- `GET /invoice-sessions/:id` - Get session

**Invoices**
- `GET /invoices` - List (filter by status, customer, date)
- `POST /invoices` - Create with items
- `GET /invoices/:id` - Get with items + payment summary
- `PUT /invoices/:id` - Update (draft only)
- `DELETE /invoices/:id` - Cancel
- `PUT /invoices/:id/status` - Update status
- `POST /invoices/:id/send` - Mark sent + WhatsApp
- `POST /invoices/:id/whatsapp` - Send via WhatsApp
- `GET /invoices/:id/pdf` - Download PDF

**Payments**
- `GET /payments` - List (paginated)
- `POST /payments` - Record payment
- `GET /payments/:id` - Get by ID
- `DELETE /payments/:id` - Delete (admin only)

**Templates**
- `GET /templates` - List
- `POST /templates` - Create
- `GET /templates/:id` - Get by ID
- `PUT /templates/:id` - Update
- `DELETE /templates/:id` - Delete
- `PUT /templates/:id/default` - Set as default

**Settings**
- `GET /settings` - Get organisation settings
- `PUT /settings` - Update settings

**Audit**
- `GET /audit` - List audit logs (admin only)

### SuperAdmin Routes (Base: `/superadmin`)

**Authentication**
- `POST /auth/login` - SuperAdmin login
- `POST /auth/refresh` - Refresh token
- `POST /auth/logout` - Logout

**Organisations**
- `GET /organisations` - List all + usage stats
- `POST /organisations` - Create/onboard new org
- `GET /organisations/:id` - Get details
- `PUT /organisations/:id` - Update
- `PUT /organisations/:id/status` - Suspend/reactivate
- `DELETE /organisations/:id` - Delete + all data (GDPR)
- `PUT /organisations/:id/plan` - Change plan
- `GET /organisations/:id/users` - List users in org
- `GET /organisations/:id/audit` - Audit log for org

**Impersonation**
- `POST /organisations/:id/impersonate` - Impersonate user
- `DELETE /impersonate/:sessionId` - End session

**Plans**
- `GET /plans` - List all plans
- `POST /plans` - Create plan
- `PUT /plans/:id` - Update plan
- `DELETE /plans/:id` - Deactivate plan

**Users**
- `GET /users` - List all users across orgs
- `PUT /users/:id/status` - Suspend/reactivate
- `POST /users/:id/reset-password` - Force reset

**Platform**
- `GET /audit` - Global audit log
- `GET /metrics` - Platform metrics (MRR, orgs, users)
- `GET /config` - Get platform config
- `PUT /config` - Update config
- `POST /maintenance` - Toggle maintenance mode

## Security Features

- **JWT RS256** (asymmetric) - separate keys for org/superadmin
- **Bcrypt** password hashing (cost 12)
- **Account lockout** after 5 failed attempts (15 min)
- **Rate limiting**: 10 req/min (auth), 300 req/min (API)
- **IP allowlist** for SuperAdmin access
- **Security headers**: HSTS, CSP, X-Frame-Options, etc.
- **CORS** with strict origin whitelist
- **Max body size**: 1MB JSON, 2MB uploads

## Business Logic

### Invoice Number Generation (Atomic)
- Format: `{prefix}-{year}-{sequence}`
- Example: `INV-2025-0001`
- Uses atomic UPDATE...RETURNING to prevent duplicates

### Service Auto-Fill
- When adding invoice item with service_id
- Copies description, unit_price, tax_rate
- Values are copied (not referenced) for historical accuracy

### Auto Customer Creation
- On invoice creation, can provide customer object
- Searches by email/phone first
- Creates new customer if not found
- All in same transaction

### Invoice Total Calculation
```
Per item:
  tax_amount = quantity × unit_price × (tax_rate / 100)
  line_total = (quantity × unit_price) + tax_amount

Invoice:
  subtotal = SUM(quantity × unit_price)
  tax_amount = SUM(item.tax_amount)
  total = subtotal + tax_amount - discount_amount
```

### Auto Status on Payment
- After recording payment, checks total paid
- If total_paid >= invoice.total → status = 'paid'
- Partial payments keep status as 'sent'

### Async WhatsApp Sending
- Returns 202 Accepted immediately
- Sends in background goroutine
- Updates whatsapp_logs table with result

### Plan Limit Enforcement
- Checks before create operations
- Invoices: max per month
- Customers: max active
- Users: max active

### Impersonation
- SuperAdmin can impersonate org users
- Requires written reason
- Issues short-lived JWT (1 hour)
- All actions logged with actor_type='superadmin'

## Project Structure

```
invoice-backend/
├── cmd/server/              # Entry point
├── internal/
│   ├── config/              # Viper configuration
│   ├── constants/           # Roles, statuses, errors
│   ├── db/
│   │   ├── migrations/      # Goose SQL migrations (19 files)
│   │   ├── queries/         # SQL for sqlc
│   │   └── sqlc/            # Generated code
│   ├── middleware/          # HTTP middleware
│   ├── modules/             # Feature modules
│   │   ├── superadmin/      # SuperAdmin features
│   │   ├── auth/            # Org user auth
│   │   ├── customers/       # Customer management
│   │   ├── services/        # Service catalog
│   │   ├── invoices/        # Invoice management
│   │   ├── payments/        # Payment tracking
│   │   ├── templates/       # Invoice templates
│   │   ├── whatsapp/        # WhatsApp integration
│   │   ├── settings/        # Org settings
│   │   └── audit/           # Audit logs
│   └── utils/               # Utilities
├── scripts/                 # Setup scripts
│   ├── run_migrations.ps1   # Migration script (Windows)
│   ├── run_migrations.sh    # Migration script (Linux/Mac)
│   ├── create_superadmin.ps1
│   └── create_superadmin.sh
├── uploads/                 # File uploads
├── tests/                   # Tests
└── keys/                    # JWT keys (gitignored)
```

## Database Tables (19 Total)

After running migrations, these tables will be created:

1. **super_admins** - Platform administrators
2. **plans** - Subscription plans
3. **organisations** - Tenant organisations
4. **organisation_subscriptions** - Org plan subscriptions
5. **users** - Organisation users
6. **refresh_tokens** - User refresh tokens
7. **super_refresh_tokens** - SuperAdmin refresh tokens
8. **impersonation_sessions** - SuperAdmin impersonation tracking
9. **customers** - Organisation customers
10. **services** - Service catalog
11. **invoice_sessions** - Invoice number sequences
12. **invoices** - Customer invoices
13. **invoice_items** - Invoice line items
14. **payments** - Invoice payments
15. **templates** - Invoice templates
16. **whatsapp_logs** - WhatsApp message logs
17. **settings** - Organisation settings
18. **audit_logs** - Organisation audit logs
19. **super_audit_logs** - SuperAdmin audit logs

All tables use **UUID** as primary keys for better distributed system support.

## Documentation

- **START_HERE.md** - Quick start guide
- **DATABASE_SETUP.md** - Complete database setup guide
- **API_ENDPOINTS.md** - All API endpoints (46+)
- **SUPERADMIN_SETUP.md** - SuperAdmin creation guide
- **COMPLETE_SETUP_GUIDE.md** - Detailed setup instructions

## License

Proprietary - All rights reserved
