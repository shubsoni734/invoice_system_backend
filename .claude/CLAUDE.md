# CLAUDE.md — Invoice Backend Project Instructions

This file contains everything Claude needs to understand, maintain, and extend this codebase. Read this file fully before making any changes.

---

## Project Overview

**Invoice Management System** — A multi-tenant SaaS backend built in Go. Each customer is an **Organisation**. A **SuperAdmin** manages organisations from a separate admin panel. Org users manage their own invoices, customers, services, payments, reports, etc.

- **Language**: Go 1.22+
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL 16 (hosted on Neon)
- **Migrations**: Goose (SQL files in `internal/pkg/db/migrations/`)
- **Query Layer**: sqlc (type-safe generated Go code from SQL)
- **Config**: Viper (reads `.env` files)
- **Logging**: Uber Zap (structured JSON in production)
- **JWT**: golang-jwt/jwt v5, HS256 (two separate managers: org + superadmin)
- **Password**: bcrypt cost 12
- **Deployment**: Docker + Railway (`railway.json`, `Dockerfile` present)

---

## Repository Structure

```
.
├── cmd/api/main.go                        # Entry point — wires everything together
├── internal/
│   ├── app/routes.go                      # Top-level route registration
│   ├── config/config.go                   # Viper env config loading
│   ├── domain/                            # One folder per business domain
│   │   ├── auth/                          # Org user auth (login, logout, forgot/reset password)
│   │   ├── common/                        # Shared endpoints (e.g. test-email)
│   │   ├── customers/                     # Customer CRUD
│   │   ├── invoices/                      # Invoice CRUD + cancel
│   │   ├── invoicesessions/               # Invoice numbering sessions
│   │   ├── payments/                      # Payment recording
│   │   ├── pdf/                           # Invoice PDF generation (gofpdf)
│   │   ├── reports/                       # Daily/monthly/revenue/customer reports
│   │   ├── roles/                         # RBAC role management
│   │   ├── services/                      # Service/product catalog
│   │   ├── settings/                      # Org settings (upsert)
│   │   ├── superadmin/
│   │   │   ├── auth/                      # SuperAdmin login/logout/create
│   │   │   ├── organisations/             # Create/list/get org, apply subscription
│   │   │   └── users/                     # Create org users, set status
│   │   ├── templates/                     # Invoice HTML templates
│   │   ├── users/                         # Org user management (org-side)
│   │   └── whatsapp/                      # WhatsApp message sending + logs
│   ├── pkg/
│   │   ├── db/migrations/                 # Goose SQL migrations (numbered 001–021)
│   │   ├── email/email.go                 # SMTP email client (Brevo/Sendinblue)
│   │   ├── middleware/                    # Auth, CORS, rate limit, RBAC, tenant, recovery, etc.
│   │   ├── response/response.go           # Standardised JSON response helpers
│   │   └── utils/                         # JWT, bcrypt, pagination, sanitize, invoice number
│   └── shared/constants/                  # Roles, statuses, error sentinels, context keys
├── sqlc.yaml                              # sqlc config — one entry per domain
├── Makefile                               # run, build, migrate-up/down, sqlc
├── Dockerfile                             # Multi-stage Alpine build
├── railway.json                           # Railway deployment config
└── .gitignore                             # Ignores .env, keys/, bin/, uploads content
```

Each domain follows this pattern:
```
domain/
  handler.go        # HTTP handlers (request parsing, DB calls, response)
  routes.go         # Route registration (called from app/routes.go)
  queries.sql       # Raw SQL (sqlc source)
  sqlc/
    db.go           # DBTX interface + Queries struct (generated)
    models.go       # Go structs for all DB tables (generated)
    queries.sql.go  # Type-safe query functions (generated)
```

---

## Environment Variables

Required variables (set in `.env.development` or `.env.production`):

```env
# Server
SERVER_PORT=8080
ENVIRONMENT=development          # or production
ALLOWED_ORIGINS=http://localhost:5173
FRONTEND_URL=http://localhost:5173

# Database
DATABASE_URL=postgres://...      # REQUIRED — Neon connection string

# JWT — two separate secrets
ORG_JWT_SECRET=...               # REQUIRED — for org user tokens
SA_JWT_SECRET=...                # REQUIRED — for superadmin tokens
ORG_ACCESS_TOKEN_EXPIRY=15m
SA_ACCESS_TOKEN_EXPIRY=15m

# Rate limiting
RATE_LIMIT_AUTH_RPM=10
RATE_LIMIT_API_RPM=300

# Email (Brevo SMTP)
SMTP_HOST=smtp-relay.brevo.com
SMTP_PORT=587
BRAVO_SMTP_USER=...              # Brevo SMTP username
BRAVO_KEY_SECRET=...             # Brevo SMTP key — NEVER commit this

# WhatsApp (optional)
WHATSAPP_API_URL=
WHATSAPP_API_KEY=

# SuperAdmin IP allowlist (comma-separated, empty = allow all)
SA_IP_ALLOWLIST=
```

**CRITICAL**: The `BRAVO_KEY_SECRET` (Sendinblue/Brevo SMTP key) must NEVER be hardcoded. GitHub push protection will block the push. Always use env vars.

---

## Authentication Architecture

### Two JWT Managers

| Manager    | Env Secret      | Protects              | Claims field   |
|------------|----------------|-----------------------|----------------|
| `orgJWT`   | `ORG_JWT_SECRET` | `/api/v1/*` routes  | `UserID`, `OrgID`, `Role` |
| `superJWT` | `SA_JWT_SECRET`  | `/superadmin/*` routes | `SuperAdminID`, `Role` |

### Claims struct (`internal/pkg/utils/jwt.go`)
```go
type Claims struct {
    UserID               string  // org user UUID
    OrgID                string  // organisation UUID
    Role                 string  // role name
    SuperAdminID         string  // set for superadmin tokens only
    ImpersonatedBy       string  // set during impersonation
    ImpersonationSession string
    jwt.RegisteredClaims
}
```

### Middleware Chain for `/api/v1/*`
1. `RateLimit(apiRateLimiter)` — 300 req/min per IP
2. `Auth(orgJWT)` — verifies Bearer token, sets ctx keys
3. `Tenant(db)` — verifies org status = 'active'

### Middleware Chain for `/superadmin/*`
1. `RateLimit(authRateLimiter)` — 10 req/min per IP
2. `SuperAuth(superJWT, ipAllowlist)` — verifies Bearer token + optional IP check

### Context Keys (`internal/shared/constants/roles.go`)
```go
CtxUserID     = "user_id"
CtxOrgID      = "organisation_id"
CtxUserRole   = "user_role"
CtxSuperAdminID = "super_admin_id"
```

---

## Database Schema (19 Tables)

Migrations are in `internal/pkg/db/migrations/` numbered 001–021.

| Migration | Table |
|-----------|-------|
| 001 | `super_admins` — email, password_hash, role, failed_attempts, locked_until |
| 002 | `plans` — name, pricing, limits (max_users, max_customers, max_invoices_per_month, etc.) |
| 003 | `organisations` — name, slug, email, status, created_by_super_admin_id |
| 004 | `organisation_subscriptions` — org→plan, status, period dates |
| 005 | `users` — organisation_id, email, password_hash, name, role, failed_attempts |
| 006 | `refresh_tokens` — user_id, token_hash, expires_at, revoked_at |
| 007 | `super_refresh_tokens` — super_admin_id, token_hash, expires_at, revoked_at |
| 008 | `impersonation_sessions` — super_admin_id, target_org_id, target_user_id |
| 009 | `customers` — org scoped, name, email, phone, address, tax_number, is_active |
| 010 | `services` — org scoped, name, unit_price, tax_rate, unit |
| 011 | `invoice_sessions` — org+year+prefix, current_sequence (for invoice numbering) |
| 012 | `invoices` — org scoped, customer_id, session_id, status, dates, totals |
| 013 | `invoice_items` — invoice_id, service_id, qty, unit_price, tax, line_total |
| 014 | `payments` — invoice_id, amount, method, payment_date, recorded_by |
| 015 | `templates` — org scoped, name, html_content, is_default |
| 016 | `whatsapp_logs` — invoice_id, recipient_phone, message, status |
| 017 | `settings` — org scoped (UNIQUE), business info, currency, invoice_prefix, defaults |
| 018 | `audit_logs` — org scoped, actor, action, resource, old/new JSONB |
| 019 | `super_audit_logs` — super_admin actions |
| 020 | `roles` — org scoped, name, description, is_system; also adds role_id to users |
| 021 | `password_resets` — user_id, token_hash, expires_at |

**All tables use UUID primary keys** (`uuid_generate_v4()`).

**Account lockout**: After 5 failed login attempts (`failed_attempts >= 4` triggers lock in the SQL `CASE` statement), account is locked for 15 minutes.

---

## sqlc Code Generation

`sqlc.yaml` has one entry per domain. After editing any `queries.sql` file, run:

```bash
make sqlc
# or
sqlc generate
```

Generated files (`db.go`, `models.go`, `queries.sql.go`) are committed to the repo — **never edit them manually**.

The `inet` PostgreSQL type maps to `*netip.Addr` in Go (not `net.IP`) — this is set in overrides per domain.

---

## API Routes

### Public (no auth)
```
GET  /health
GET  /ready
POST /api/v1/auth/login
POST /api/v1/auth/forgot-password
POST /api/v1/auth/reset-password
POST /api/v1/common/test-email
POST /superadmin/auth/create
POST /superadmin/auth/login
```

### Org API (Bearer token required — `orgJWT`)
```
GET  /api/v1/auth/me
POST /api/v1/auth/logout

GET/POST/PUT/DELETE  /api/v1/roles
GET/POST/PUT         /api/v1/users
PUT                  /api/v1/users/:id/status

GET/PUT              /api/v1/settings
GET/POST/PUT/DELETE  /api/v1/customers
GET/POST/PUT/DELETE  /api/v1/services
GET/POST             /api/v1/invoice-sessions
GET/POST/PUT/DELETE  /api/v1/templates
GET/POST             /api/v1/invoices
GET                  /api/v1/invoices/:id
PUT                  /api/v1/invoices/:id/cancel
GET                  /api/v1/invoices/:id/pdf
GET                  /api/v1/invoices/:id/payments
POST                 /api/v1/payments
DELETE               /api/v1/payments/:id

GET  /api/v1/reports/daily?date=YYYY-MM-DD
GET  /api/v1/reports/monthly?year=YYYY&month=M
GET  /api/v1/reports/customer/:id
GET  /api/v1/reports/revenue

GET/POST /api/v1/whatsapp/logs
POST     /api/v1/whatsapp/send
```

### SuperAdmin API (Bearer token required — `superJWT`)
```
GET  /superadmin/auth/me
POST /superadmin/auth/logout

POST /superadmin/organisations
GET  /superadmin/organisations
GET  /superadmin/organisations/:id
POST /superadmin/organisations/:id/subscription

GET  /superadmin/organisations/:id/users
POST /superadmin/organisations/:id/users
PUT  /superadmin/users/:id/status
```

---

## Response Format

All responses use `internal/pkg/response/response.go`:

```json
{
  "success": true,
  "message": "...",
  "data": { ... },
  "meta": { "page": 1, "per_page": 20, "total": 100, "total_pages": 5 },
  "request_id": "uuid"
}
```

Helper functions:
- `response.Success(c, statusCode, message, data)`
- `response.SuccessWithMeta(c, statusCode, message, data, &Meta{...})`
- `response.Error(c, statusCode, message)`
- `response.ValidationError(c, statusCode, []FieldError{...})`

---

## Key Patterns & Conventions

### Handler Pattern
Every handler follows this structure:
```go
func (h *Handler) DoSomething(c *gin.Context) {
    // 1. Extract org/user ID from context
    orgID, err := getOrgID(c)
    if err != nil { response.Error(c, 401, "..."); return }

    // 2. Parse & validate request body
    var req SomeRequest
    if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, 400, "..."); return }

    // 3. Call sqlc query
    result, err := h.q.SomeQuery(context.Background(), params)
    if err != nil { response.Error(c, 500, "..."); return }

    // 4. Return response
    response.Success(c, 200, "Done", result)
}
```

### pgtype.Numeric Conversion
When converting `float64` to `pgtype.Numeric` for DB writes, use the pattern:
```go
func floatToNumeric(f float64) pgtype.Numeric {
    var num pgtype.Numeric
    strValue := fmt.Sprintf("%v", f)
    err := num.Scan(strValue)
    if err != nil { num.Valid = false } else { num.Valid = true }
    return num
}
```

### pgtype.UUID for nullable UUIDs
```go
// Set a UUID
pgtype.UUID{Bytes: someUUID, Valid: true}

// Null UUID
pgtype.UUID{Valid: false}
```

### Pagination
Use `utils.GetPaginationParams(c)` which reads `?page=1&per_page=20` query params (defaults: page=1, per_page=20, max=100).

### Context Keys
Always use `c.GetString(constants.CtxOrgID)` — not raw strings — when reading context values.

### Role System
Two role systems coexist:
1. **String role** on `users.role` column: `"admin"`, `"manager"`, `"viewer"`
2. **Role table** (`roles`) with UUID FK `users.role_id`: org-defined custom roles

When creating the first user for an org, always assign them as `admin` and create the "Admin" system role if it doesn't exist.

---

## Security Rules

1. **Tokens**: Never commit JWT secrets or SMTP keys. GitHub push protection will block it.
2. **Password hashing**: Always use `utils.HashPassword()` (bcrypt cost 12). Never store plain passwords.
3. **Token hashing**: Refresh tokens are stored as SHA-256 hashes (`utils.HashToken()`).
4. **Account lockout**: `failed_attempts >= 4` → lock for 15 minutes. Reset on successful login.
5. **Multi-tenancy**: Every query that touches org data MUST filter by `organisation_id`. Never return data across orgs.
6. **SuperAdmin IP allowlist**: Configurable via `SA_IP_ALLOWLIST`. If set, only those IPs can reach `/superadmin/*`.
7. **CORS**: Strict origin whitelist from `ALLOWED_ORIGINS` env var.
8. **Security headers**: HSTS, X-Frame-Options DENY, X-Content-Type-Options nosniff, Referrer-Policy.

---

## Email System

Client: `internal/pkg/email/email.go` using standard `net/smtp` with STARTTLS.

Provider: Brevo (formerly Sendinblue) SMTP relay.

Two email types implemented:
- `SendWelcomeEmail(toEmail, orgName, password)` — sent when superadmin creates an org user
- `SendForgotPasswordEmail(toEmail, token, frontendURL)` — password reset link (10 min expiry)

Reset link format: `{FRONTEND_URL}/new-password?refreshtoken={token}`

**If email sending fails, the operation fails** (no silent swallowing).

---

## PDF Generation

`internal/domain/pdf/` uses `github.com/jung-kurt/gofpdf`.

`buildInvoicePDF()` in `builder.go` takes: invoice, items, customer, settings and returns `[]byte`.

The PDF handler (`handler.go`) fetches all required data then calls the builder. It sets `Content-Disposition: attachment` so the browser downloads the file.

Color palette (indigo theme):
- Header: `#1E40AF` (indigo-800)
- Accent: `#6366F1` (indigo-500)
- Light bg: `#EEF2FF` (indigo-50)

---

## Invoice Numbering

`invoice_sessions` table tracks sequences per org/year/prefix.

`utils.FormatInvoiceNumber(prefix, year, sequence)` formats as `{PREFIX}-{YEAR}-{SEQUENCE:04d}`.

Example: `INV-2025-0001`

The `UpdateInvoiceSessionSequence` query atomically increments `current_sequence`.

---

## WhatsApp Integration

`internal/domain/whatsapp/` — sends messages via an external WhatsApp API (configurable via `WHATSAPP_API_URL` + `WHATSAPP_API_KEY`).

If API URL/key are not configured, the message is still logged with status `"sent"` (mock mode).

If the external API returns HTTP 4xx/5xx, status is logged as `"failed"` with the error message.

---

## Plan Limits Middleware

`internal/pkg/middleware/plan_limit.go` runs on POST requests only. It checks:
- `/api/v1/invoices` → enforces `max_invoices_per_month`
- `/api/v1/customers` → enforces `max_customers`

If limit is exceeded → HTTP 403.

---

## Makefile Commands

```bash
make run            # go run ./cmd/api/main.go
make build          # builds to bin/invoice-backend
make tidy           # go mod tidy
make sqlc           # regenerate sqlc code
make migrate-up     # run pending migrations
make migrate-down   # roll back last migration
make migrate-status # show migration status
```

The `DB_URL` in Makefile is extracted from `.env.development` automatically.

---

## Adding a New Domain

Follow this exact pattern:

1. **Create migration**: `internal/pkg/db/migrations/NNN_create_X.sql`
2. **Write SQL queries**: `internal/domain/X/queries.sql`
3. **Add sqlc config**: Add entry to `sqlc.yaml`
4. **Generate code**: `make sqlc`
5. **Write handler**: `internal/domain/X/handler.go`
6. **Register routes**: `internal/domain/X/routes.go`
7. **Wire in app**: Add `X.RegisterRoutes(...)` call in `internal/app/routes.go`

---

## Adding a New Query

1. Add SQL to the domain's `queries.sql` with `-- name: QueryName :one/:many/:exec`
2. Run `make sqlc` to regenerate
3. Use the generated function in the handler

**Never edit** `sqlc/db.go`, `sqlc/models.go`, or `sqlc/queries.sql.go` manually.

---

## Known Issues / Technical Debt

1. **No transactions on invoice creation**: `CreateInvoice` + multiple `CreateInvoiceItem` calls are not wrapped in a DB transaction. If item insertion fails, the invoice header exists without items. Fix: use `pgxpool.BeginTx` and `q.WithTx(tx)`.

2. **Invoice handler has unused `db *pgx.Conn` field**: `invoices/handler.go` declares `db *pgx.Conn` but never uses it. Safe to remove.

3. **SuperAdmin org creation doesn't set `created_by_super_admin_id`**: `organisations/handler.go` passes `pgtype.UUID{Valid: false}`. Should read from `CtxSuperAdminID`.

4. **No refresh token rotation**: The refresh token flow stores tokens but doesn't implement the `/auth/refresh` endpoint yet. Tokens expire after 7 days with no renewal mechanism.

5. **`common/test-email` is public**: The test email endpoint is not behind auth, so anyone can trigger emails. Should be moved behind auth or removed in production.

6. **Customers `GetCustomers` LIMIT/OFFSET order mismatch**: The generated SQL uses `LIMIT $4 OFFSET $3` (positional swap). The sqlc params pass them correctly (`Offset` as $3, `PerPage` as $4) — this is intentional but confusing.

---

## Development Workflow

```bash
# First time setup
cp .env.development .env
# Edit .env with real DATABASE_URL, ORG_JWT_SECRET, SA_JWT_SECRET

# Install tools
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Run migrations
make migrate-up

# Start server
make run
```

Server runs on `http://localhost:8080`. Use the included Postman collection (`Invoice_Management_System.postman_collection.json`) for testing all endpoints.

---

## Deployment (Railway)

The `railway.json` uses `DOCKERFILE` builder with `ON_FAILURE` restart policy (max 3 retries).

The `Dockerfile` is a two-stage build: Go Alpine builder → minimal Alpine runtime with `ca-certificates` and `tzdata`.

Set all required env vars in the Railway dashboard. The `DATABASE_URL` should point to Neon PostgreSQL.

---

## GitHub Push Protection

The repo has GitHub push protection enabled. It will block pushes containing secrets.

The Brevo SMTP key (`BRAVO_KEY_SECRET`) was previously caught at path `internal/pkg/email/email.go:13`. **Never hardcode credentials**. Always use environment variables loaded via Viper.

If you get blocked:
1. Remove the secret from the commit: `git rebase -i` and edit the commit
2. Or use `git filter-branch` / BFG to purge from history
3. Rotate the leaked credential immediately

---

## Module Path

```
github.com/your-org/invoice-backend
```

All internal imports use this prefix. If the repo is forked/renamed, update `go.mod` and all import paths.