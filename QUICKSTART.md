# Quick Start Guide

## What's Been Created

✅ **Core Files**
- `go.mod` - Go module with all dependencies
- `.env.example` - Environment template with Neon DB URL
- `sqlc.yaml` - SQLC configuration
- `.gitignore` - Proper exclusions
- `README.md` - Complete documentation

✅ **Database (19 migrations)**
- All tables created in `internal/db/migrations/`
- Ready to run with goose

✅ **Constants**
- `internal/constants/roles.go` - Context keys, roles
- `internal/constants/status.go` - All status constants
- `internal/constants/errors.go` - Typed errors

✅ **Utilities**
- `internal/utils/jwt.go` - RS256 JWT manager
- `internal/utils/hash.go` - Bcrypt password hashing
- `internal/utils/response.go` - Standard JSON responses
- `internal/utils/pagination.go` - Pagination helpers
- `internal/utils/invoice_number.go` - Invoice number formatter
- `internal/utils/sanitize.go` - Input sanitization

✅ **Configuration**
- `internal/config/config.go` - Viper config loader

✅ **Middleware (13 files)**
- Recovery, RequestID, Logger, Security Headers
- CORS, Rate Limiting
- Auth, SuperAuth, RBAC, SuperRBAC
- Tenant, Plan Limit, Error Handler

✅ **Main Application**
- `cmd/server/main.go` - Complete server with graceful shutdown

## Next Steps

### 1. Install CLI Tools
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### 2. Generate JWT Keys
```bash
mkdir keys
openssl genrsa -out keys/org_private.pem 2048
openssl rsa -in keys/org_private.pem -pubout -out keys/org_public.pem
openssl genrsa -out keys/sa_private.pem 2048
openssl rsa -in keys/sa_private.pem -pubout -out keys/sa_public.pem
```

### 3. Setup Environment
```bash
cp .env.example .env
# .env already has your Neon database URL
```

### 4. Download Dependencies
```bash
go mod download
```

### 5. Run Migrations
```bash
$env:DATABASE_URL="postgresql://neondb_owner:npg_GRkOWz9Hf6rI@ep-mute-mud-adi7b3r8-pooler.c-2.us-east-1.aws.neon.tech/neondb?sslmode=require"
goose -dir internal/db/migrations postgres $env:DATABASE_URL up
```

### 6. Create SQLC Queries
You need to create query files in `internal/db/queries/` for each table.
Example structure already provided in the original files.

### 7. Generate SQLC Code
```bash
sqlc generate
```

### 8. Run Server
```bash
go run cmd/server/main.go
```

Server will start on `http://localhost:8080`

## Test Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready
```

## What's Missing (To Complete)

1. **SQLC Query Files** - Create SQL queries in `internal/db/queries/` for:
   - super_admins.sql
   - plans.sql
   - organisations.sql
   - subscriptions.sql
   - users.sql
   - refresh_tokens.sql
   - customers.sql
   - services.sql
   - invoice_sessions.sql
   - invoices.sql
   - invoice_items.sql
   - payments.sql
   - templates.sql
   - whatsapp_logs.sql
   - settings.sql
   - audit_logs.sql
   - super_audit_logs.sql

2. **Module Implementation** - Create handlers/services/routes for:
   - `internal/modules/auth/` - Org user authentication
   - `internal/modules/customers/` - Customer management
   - `internal/modules/services/` - Service catalog
   - `internal/modules/invoices/` - Invoice management
   - `internal/modules/payments/` - Payment tracking
   - `internal/modules/templates/` - Template management
   - `internal/modules/settings/` - Org settings
   - `internal/modules/superadmin/auth/` - SuperAdmin auth
   - `internal/modules/superadmin/organisations/` - Org management
   - `internal/modules/superadmin/plans/` - Plan management

## Project Structure

```
invoice-backend/
├── cmd/server/main.go          ✅ Created
├── internal/
│   ├── config/config.go        ✅ Created
│   ├── constants/              ✅ All created
│   ├── db/
│   │   ├── migrations/         ✅ All 19 created
│   │   ├── queries/            ⚠️  Need to create
│   │   └── sqlc/               ⚠️  Generated after queries
│   ├── middleware/             ✅ All 13 created
│   ├── modules/                ⚠️  Need to implement
│   └── utils/                  ✅ All created
├── uploads/                    ✅ Created
├── tests/                      ✅ Created
├── go.mod                      ✅ Created
├── .env.example                ✅ Created
├── sqlc.yaml                   ✅ Created
└── README.md                   ✅ Created
```

## Tips

- The database URL in `.env.example` is already configured for your Neon database
- All middleware is production-ready with proper error handling
- JWT uses RS256 (asymmetric) - never commit keys to git
- Rate limiting is configured: 10 req/min for auth, 300 req/min for API
- All migrations follow goose format with Up/Down sections

## Need Help?

Check `README.md` for complete API documentation and all available commands.
