# Invoice Backend — AI Generation Prompt

> **How to use this file:**
> Paste the entire contents into your AI IDE (Cursor, Windsurf, GitHub Copilot Workspace, etc.) and say:
> _"Generate a complete Go backend project based on this specification. Follow every section exactly. Do not skip any module, middleware, route, database table, or file."_

---

## 1. PROJECT OVERVIEW

| Field | Value |
|---|---|
| Project Name | `invoice-backend` |
| Type | Multi-Tenant SaaS — Invoice Management Platform |
| Language | Go 1.22+ |
| Architecture | REST API, Modular, Feature-based |
| Database | PostgreSQL 16 |
| Authentication | JWT RS256 (asymmetric keys) |
| Run Command | `go run cmd/server/main.go` |

This backend serves as a platform where the **platform owner (SuperAdmin)** manages multiple businesses (organisations), and each business manages their own customers, services, invoices, payments, and templates.

---

## 2. TECH STACK — USE EXACTLY THESE PACKAGES

```
Language        : Go 1.22+
Web Framework   : github.com/gin-gonic/gin
DB Driver       : github.com/jackc/pgx/v5
DB Pool         : github.com/jackc/pgx/v5/pgxpool
Query Generator : github.com/sqlc-dev/sqlc             (sqlc CLI tool)
Migrations      : github.com/pressly/goose/v3          (goose CLI tool)
Config          : github.com/spf13/viper
Logging         : go.uber.org/zap
Validation      : github.com/go-playground/validator/v10
JWT             : github.com/golang-jwt/jwt/v5          — RS256 ONLY, never HS256
Password Hash   : golang.org/x/crypto/bcrypt            — cost 12
Rate Limiting   : golang.org/x/time/rate
UUID            : github.com/google/uuid
Metrics         : github.com/prometheus/client_golang
Testing         : github.com/stretchr/testify
```

### go.mod — generate with these dependencies

```
module github.com/your-org/invoice-backend

go 1.22

require (
    github.com/gin-gonic/gin v1.10.0
    github.com/jackc/pgx/v5 v5.6.0
    github.com/pressly/goose/v3 v3.20.0
    github.com/spf13/viper v1.19.0
    go.uber.org/zap v1.27.0
    github.com/go-playground/validator/v10 v10.22.0
    github.com/golang-jwt/jwt/v5 v5.2.1
    golang.org/x/crypto v0.24.0
    golang.org/x/time v0.5.0
    github.com/google/uuid v1.6.0
    github.com/prometheus/client_golang v1.19.1
    go.opentelemetry.io/otel v1.27.0
    github.com/stretchr/testify v1.9.0
)
```

---

## 3. HOW TO RUN — PURE GO COMMANDS (NO MAKE, NO DOCKER)

> Generate a `README.md` that documents these exact commands. No Makefile. No Docker. Everything runs with standard Go and CLI tools.

### Install required CLI tools (one time only)

```bash
# Install goose migration tool
go install github.com/pressly/goose/v3/cmd/goose@latest

# Install sqlc query generator
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Install linter (optional)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Generate JWT signing keys (one time only)

```bash
mkdir -p keys

# Organisation user JWT keys
openssl genrsa -out keys/org_private.pem 2048
openssl rsa -in keys/org_private.pem -pubout -out keys/org_public.pem

# SuperAdmin JWT keys
openssl genrsa -out keys/sa_private.pem 2048
openssl rsa -in keys/sa_private.pem -pubout -out keys/sa_public.pem
```

### First-time setup

```bash
# 1. Clone and enter project
git clone <repo-url>
cd invoice-backend

# 2. Copy and fill environment file
cp .env.example .env

# 3. Download Go dependencies
go mod download

# 4. Generate JWT keys (see above)

# 5. Run all database migrations
goose -dir internal/db/migrations postgres "$DATABASE_URL" up

# 6. Generate type-safe query code
sqlc generate

# 7. Start the server
go run cmd/server/main.go
```

### Day-to-day commands

```bash
# Run the server
go run cmd/server/main.go

# Build a binary
go build -o bin/server cmd/server/main.go

# Run the built binary
./bin/server

# Run all tests
go test ./... -v -count=1

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Apply all pending migrations
goose -dir internal/db/migrations postgres "$DATABASE_URL" up

# Roll back last migration
goose -dir internal/db/migrations postgres "$DATABASE_URL" down

# Check migration status
goose -dir internal/db/migrations postgres "$DATABASE_URL" status

# Regenerate sqlc query code (run after any query file change)
sqlc generate

# Run linter
golangci-lint run ./...
```

---

## 4. FOLDER STRUCTURE — GENERATE EXACTLY THIS

```
invoice-backend/
│
├── cmd/
│   └── server/
│       └── main.go                         # Entry point — server init, route wiring, graceful shutdown
│
├── internal/
│   │
│   ├── config/
│   │   └── config.go                       # Viper config loader — reads .env into Config struct
│   │
│   ├── db/
│   │   │
│   │   ├── migrations/                     # Goose SQL migration files — one file per change
│   │   │   ├── 001_create_super_admins.sql
│   │   │   ├── 002_create_plans.sql
│   │   │   ├── 003_create_organisations.sql
│   │   │   ├── 004_create_organisation_subscriptions.sql
│   │   │   ├── 005_create_users.sql
│   │   │   ├── 006_create_refresh_tokens.sql
│   │   │   ├── 007_create_super_refresh_tokens.sql
│   │   │   ├── 008_create_impersonation_sessions.sql
│   │   │   ├── 009_create_customers.sql
│   │   │   ├── 010_create_services.sql
│   │   │   ├── 011_create_invoice_sessions.sql
│   │   │   ├── 012_create_invoices.sql
│   │   │   ├── 013_create_invoice_items.sql
│   │   │   ├── 014_create_payments.sql
│   │   │   ├── 015_create_templates.sql
│   │   │   ├── 016_create_whatsapp_logs.sql
│   │   │   ├── 017_create_settings.sql
│   │   │   ├── 018_create_audit_logs.sql
│   │   │   └── 019_create_super_audit_logs.sql
│   │   │
│   │   ├── queries/                        # Raw SQL query files — consumed by sqlc
│   │   │   ├── super_admins.sql
│   │   │   ├── plans.sql
│   │   │   ├── organisations.sql
│   │   │   ├── subscriptions.sql
│   │   │   ├── users.sql
│   │   │   ├── refresh_tokens.sql
│   │   │   ├── super_refresh_tokens.sql
│   │   │   ├── impersonation_sessions.sql
│   │   │   ├── customers.sql
│   │   │   ├── services.sql
│   │   │   ├── invoice_sessions.sql
│   │   │   ├── invoices.sql
│   │   │   ├── invoice_items.sql
│   │   │   ├── payments.sql
│   │   │   ├── templates.sql
│   │   │   ├── whatsapp_logs.sql
│   │   │   ├── settings.sql
│   │   │   ├── audit_logs.sql
│   │   │   └── super_audit_logs.sql
│   │   │
│   │   └── sqlc/                           # AUTO-GENERATED by sqlc — DO NOT EDIT MANUALLY
│   │       ├── db.go
│   │       ├── models.go
│   │       └── *.sql.go
│   │
│   ├── modules/
│   │   │
│   │   ├── superadmin/                     # ═══ SUPERADMIN PLANE ═══
│   │   │   ├── auth/
│   │   │   │   ├── handler.go              # HTTP handlers (thin — parse, call service, respond)
│   │   │   │   ├── service.go              # Business logic
│   │   │   │   ├── routes.go               # Route registration
│   │   │   │   └── validation.go           # Request struct + validation tags
│   │   │   ├── organisations/
│   │   │   │   ├── handler.go
│   │   │   │   ├── service.go
│   │   │   │   └── routes.go
│   │   │   ├── plans/
│   │   │   │   ├── handler.go
│   │   │   │   ├── service.go
│   │   │   │   └── routes.go
│   │   │   ├── users/
│   │   │   │   ├── handler.go
│   │   │   │   ├── service.go
│   │   │   │   └── routes.go
│   │   │   ├── impersonation/
│   │   │   │   ├── handler.go
│   │   │   │   ├── service.go
│   │   │   │   └── routes.go
│   │   │   ├── metrics/
│   │   │   │   ├── handler.go
│   │   │   │   └── routes.go
│   │   │   └── config/
│   │   │       ├── handler.go
│   │   │       ├── service.go
│   │   │       └── routes.go
│   │   │
│   │   ├── auth/                           # ═══ ORG USER PLANE ═══
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── routes.go
│   │   │   └── validation.go
│   │   │
│   │   ├── customers/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── routes.go
│   │   │   └── validation.go
│   │   │
│   │   ├── services/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── routes.go
│   │   │   └── validation.go
│   │   │
│   │   ├── invoice_sessions/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── routes.go
│   │   │
│   │   ├── invoices/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── routes.go
│   │   │   └── validation.go
│   │   │
│   │   ├── invoice_items/
│   │   │   ├── handler.go
│   │   │   └── service.go
│   │   │
│   │   ├── payments/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── routes.go
│   │   │   └── validation.go
│   │   │
│   │   ├── templates/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── routes.go
│   │   │
│   │   ├── whatsapp/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── routes.go
│   │   │
│   │   ├── settings/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── routes.go
│   │   │
│   │   └── audit/
│   │       ├── handler.go
│   │       ├── service.go
│   │       └── routes.go
│   │
│   ├── middleware/
│   │   ├── super_auth.go                   # Verify SuperAdmin JWT, load super_admin into context
│   │   ├── super_rbac.go                   # Enforce SuperAdmin role per route
│   │   ├── auth.go                         # Verify Org user JWT, load user into context
│   │   ├── rbac.go                         # Enforce Org user role per route
│   │   ├── tenant.go                       # Load organisation, verify subscription active, inject into context
│   │   ├── plan_limit.go                   # Check plan limits before create operations
│   │   ├── ratelimit.go                    # Per-IP + per-user rate limiting
│   │   ├── cors.go                         # Strict origin whitelist — never wildcard
│   │   ├── security_headers.go             # HSTS, CSP, X-Frame-Options, nosniff, Referrer-Policy
│   │   ├── requestid.go                    # Inject unique X-Request-ID into every request
│   │   ├── logger.go                       # Structured Zap request logging (method, path, status, latency)
│   │   ├── recovery.go                     # Panic recovery — convert to structured 500 error
│   │   └── error.go                        # Central typed-error to HTTP status code mapper
│   │
│   ├── utils/
│   │   ├── response.go                     # Standard JSON success/error envelope helpers
│   │   ├── jwt.go                          # RS256 token issue / verify / refresh for both issuers
│   │   ├── hash.go                         # bcrypt password hash + compare helpers
│   │   ├── invoice_number.go               # Atomic session-scoped invoice number generator
│   │   ├── pagination.go                   # Cursor-based and offset pagination helpers
│   │   └── sanitize.go                     # Input sanitisation helpers
│   │
│   └── constants/
│       ├── roles.go                        # SuperAdmin roles + Org user roles
│       ├── status.go                       # Invoice, payment, subscription status values
│       └── errors.go                       # Typed sentinel errors + HTTP code mapping
│
├── uploads/
│   ├── logos/                              # Uploaded organisation logo images
│   └── invoices/                           # Generated invoice PDF files
│
├── tests/
│   ├── integration/                        # End-to-end API tests using httptest
│   └── unit/                               # Unit tests for service layer
│
├── keys/                                   # RS256 PEM key files — NEVER commit to git
│   ├── org_private.pem
│   ├── org_public.pem
│   ├── sa_private.pem
│   └── sa_public.pem
│
├── .env.example                            # Template env file — placeholder values only
├── .gitignore                              # Must include: .env, keys/, uploads/, bin/
├── sqlc.yaml                               # sqlc configuration
├── go.mod
├── go.sum
└── README.md                               # Setup and run instructions using go commands only
```

---

## 5. PERMISSION MODEL — THREE TIERS

```
+──────────────────────────────────────────────────────────────+
|  TIER 1 — SUPERADMIN (Platform Owner)                        |
|  Roles   : superadmin | support | finance | readonly         |
|  Scope   : ALL organisations, ALL data on the platform       |
|  Auth    : POST /superadmin/auth/login                       |
|  JWT Key : SA_JWT_PRIVATE_KEY  (separate key pair)           |
+──────────────────────────────────────────────────────────────+
|  TIER 2 — ORGANISATION ADMIN (Your customers)                |
|  Roles   : admin | manager | viewer                          |
|  Scope   : OWN organisation ONLY                             |
|  Auth    : POST /api/v1/auth/login                           |
|  JWT Key : ORG_JWT_PRIVATE_KEY (separate key pair)           |
+──────────────────────────────────────────────────────────────+
|  TIER 3 — ORGANISATION MEMBER (Future)                       |
|  Roles   : staff                                             |
|  Scope   : Limited operations within own organisation        |
+──────────────────────────────────────────────────────────────+
```

**Critical rule:** SuperAdmin JWTs and Org JWTs use **different private keys** and **different middleware chains**. A SuperAdmin token must never grant access to org routes and vice versa.

---

## 6. DATABASE SCHEMA — ALL 19 TABLES

> Generate all migration files in `internal/db/migrations/`. Every file must have `-- +goose Up` and `-- +goose Down` sections.

### 001 — super_admins
```sql
-- +goose Up
CREATE TABLE super_admins (
    id              BIGSERIAL PRIMARY KEY,
    email           TEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    role            TEXT NOT NULL DEFAULT 'superadmin',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    failed_attempts INT NOT NULL DEFAULT 0,
    locked_until    TIMESTAMPTZ,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose Down
DROP TABLE IF EXISTS super_admins;
```

### 002 — plans
```sql
-- +goose Up
CREATE TABLE plans (
    id                      BIGSERIAL PRIMARY KEY,
    name                    TEXT NOT NULL UNIQUE,
    price_monthly           NUMERIC(10,2) NOT NULL DEFAULT 0,
    price_yearly            NUMERIC(10,2) NOT NULL DEFAULT 0,
    max_users               INT NOT NULL DEFAULT 1,
    max_customers           INT NOT NULL DEFAULT 100,
    max_invoices_per_month  INT NOT NULL DEFAULT 50,
    max_storage_mb          INT NOT NULL DEFAULT 500,
    whatsapp_enabled        BOOLEAN NOT NULL DEFAULT FALSE,
    custom_templates        BOOLEAN NOT NULL DEFAULT FALSE,
    api_access              BOOLEAN NOT NULL DEFAULT FALSE,
    is_active               BOOLEAN NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose Down
DROP TABLE IF EXISTS plans;
```

### 003 — organisations
```sql
-- +goose Up
CREATE TABLE organisations (
    id                        BIGSERIAL PRIMARY KEY,
    name                      TEXT NOT NULL,
    slug                      TEXT UNIQUE NOT NULL,
    email                     TEXT,
    phone                     TEXT,
    address                   TEXT,
    logo_url                  TEXT,
    status                    TEXT NOT NULL DEFAULT 'active',
    created_by_super_admin_id BIGINT REFERENCES super_admins(id) ON DELETE SET NULL,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose Down
DROP TABLE IF EXISTS organisations;
```

### 004 — organisation_subscriptions
```sql
-- +goose Up
CREATE TABLE organisation_subscriptions (
    id                   BIGSERIAL PRIMARY KEY,
    organisation_id      BIGINT NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    plan_id              BIGINT NOT NULL REFERENCES plans(id),
    status               TEXT NOT NULL DEFAULT 'trialing',
    trial_ends_at        TIMESTAMPTZ,
    current_period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_period_end   TIMESTAMPTZ NOT NULL,
    cancelled_at         TIMESTAMPTZ,
    external_id          TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sub_org ON organisation_subscriptions(organisation_id);
-- +goose Down
DROP TABLE IF EXISTS organisation_subscriptions;
```

### 005 — users
```sql
-- +goose Up
CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    organisation_id BIGINT NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    email           TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    name            TEXT NOT NULL,
    role            TEXT NOT NULL DEFAULT 'admin',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    failed_attempts INT NOT NULL DEFAULT 0,
    locked_until    TIMESTAMPTZ,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organisation_id, email)
);
CREATE INDEX idx_users_org ON users(organisation_id);
-- +goose Down
DROP TABLE IF EXISTS users;
```

### 006 — refresh_tokens
```sql
-- +goose Up
CREATE TABLE refresh_tokens (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash   TEXT NOT NULL UNIQUE,
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked_at   TIMESTAMPTZ,
    ip_address   INET,
    user_agent   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_rt_user ON refresh_tokens(user_id);
-- +goose Down
DROP TABLE IF EXISTS refresh_tokens;
```

### 007 — super_refresh_tokens
```sql
-- +goose Up
CREATE TABLE super_refresh_tokens (
    id             BIGSERIAL PRIMARY KEY,
    super_admin_id BIGINT NOT NULL REFERENCES super_admins(id) ON DELETE CASCADE,
    token_hash     TEXT NOT NULL UNIQUE,
    expires_at     TIMESTAMPTZ NOT NULL,
    revoked_at     TIMESTAMPTZ,
    ip_address     INET,
    user_agent     TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_srt_super_admin ON super_refresh_tokens(super_admin_id);
-- +goose Down
DROP TABLE IF EXISTS super_refresh_tokens;
```

### 008 — impersonation_sessions
```sql
-- +goose Up
CREATE TABLE impersonation_sessions (
    id             BIGSERIAL PRIMARY KEY,
    super_admin_id BIGINT NOT NULL REFERENCES super_admins(id),
    target_org_id  BIGINT NOT NULL REFERENCES organisations(id),
    target_user_id BIGINT NOT NULL REFERENCES users(id),
    reason         TEXT NOT NULL,
    token_hash     TEXT NOT NULL,
    started_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at       TIMESTAMPTZ,
    ip_address     INET
);
CREATE INDEX idx_imp_super_admin ON impersonation_sessions(super_admin_id);
-- +goose Down
DROP TABLE IF EXISTS impersonation_sessions;
```

### 009 — customers
```sql
-- +goose Up
CREATE TABLE customers (
    id              BIGSERIAL PRIMARY KEY,
    organisation_id BIGINT NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    email           TEXT,
    phone           TEXT,
    address         TEXT,
    tax_number      TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_customers_org   ON customers(organisation_id);
CREATE INDEX idx_customers_email ON customers(organisation_id, email);
CREATE INDEX idx_customers_phone ON customers(organisation_id, phone);
-- +goose Down
DROP TABLE IF EXISTS customers;
```

### 010 — services
```sql
-- +goose Up
CREATE TABLE services (
    id              BIGSERIAL PRIMARY KEY,
    organisation_id BIGINT NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT,
    unit_price      NUMERIC(12,2) NOT NULL DEFAULT 0,
    tax_rate        NUMERIC(5,2) NOT NULL DEFAULT 0,
    unit            TEXT NOT NULL DEFAULT 'unit',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_services_org ON services(organisation_id);
-- +goose Down
DROP TABLE IF EXISTS services;
```

### 011 — invoice_sessions
```sql
-- +goose Up
CREATE TABLE invoice_sessions (
    id               BIGSERIAL PRIMARY KEY,
    organisation_id  BIGINT NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    year             INT NOT NULL,
    prefix           TEXT NOT NULL DEFAULT 'INV',
    current_sequence INT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organisation_id, year, prefix)
);
-- Invoice number format: {prefix}-{year}-{sequence zero-padded to 4 digits}
-- Example: INV-2025-0001
-- +goose Down
DROP TABLE IF EXISTS invoice_sessions;
```

### 012 — invoices
```sql
-- +goose Up
CREATE TABLE invoices (
    id              BIGSERIAL PRIMARY KEY,
    organisation_id BIGINT NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    customer_id     BIGINT NOT NULL REFERENCES customers(id),
    session_id      BIGINT NOT NULL REFERENCES invoice_sessions(id),
    invoice_number  TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'draft',
    issued_date     DATE NOT NULL DEFAULT CURRENT_DATE,
    due_date        DATE NOT NULL,
    subtotal        NUMERIC(12,2) NOT NULL DEFAULT 0,
    tax_amount      NUMERIC(12,2) NOT NULL DEFAULT 0,
    discount_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    total           NUMERIC(12,2) NOT NULL DEFAULT 0,
    currency        TEXT NOT NULL DEFAULT 'USD',
    notes           TEXT,
    terms           TEXT,
    template_id     BIGINT,
    created_by      BIGINT REFERENCES users(id),
    sent_at         TIMESTAMPTZ,
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organisation_id, invoice_number)
);
CREATE INDEX idx_invoices_org_status ON invoices(organisation_id, status);
CREATE INDEX idx_invoices_customer   ON invoices(customer_id);
CREATE INDEX idx_invoices_due_date   ON invoices(due_date);
CREATE INDEX idx_invoices_issued     ON invoices(issued_date);
-- +goose Down
DROP TABLE IF EXISTS invoices;
```

### 013 — invoice_items
```sql
-- +goose Up
CREATE TABLE invoice_items (
    id          BIGSERIAL PRIMARY KEY,
    invoice_id  BIGINT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    service_id  BIGINT REFERENCES services(id) ON DELETE SET NULL,
    description TEXT NOT NULL,
    quantity    NUMERIC(10,3) NOT NULL DEFAULT 1,
    unit_price  NUMERIC(12,2) NOT NULL,
    tax_rate    NUMERIC(5,2) NOT NULL DEFAULT 0,
    tax_amount  NUMERIC(12,2) NOT NULL DEFAULT 0,
    line_total  NUMERIC(12,2) NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_items_invoice ON invoice_items(invoice_id);
-- +goose Down
DROP TABLE IF EXISTS invoice_items;
```

### 014 — payments
```sql
-- +goose Up
CREATE TABLE payments (
    id              BIGSERIAL PRIMARY KEY,
    organisation_id BIGINT NOT NULL REFERENCES organisations(id),
    invoice_id      BIGINT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    amount          NUMERIC(12,2) NOT NULL,
    method          TEXT NOT NULL DEFAULT 'cash',
    reference       TEXT,
    notes           TEXT,
    payment_date    DATE NOT NULL DEFAULT CURRENT_DATE,
    recorded_by     BIGINT REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_payments_invoice ON payments(invoice_id);
CREATE INDEX idx_payments_org     ON payments(organisation_id);
-- +goose Down
DROP TABLE IF EXISTS payments;
```

### 015 — templates
```sql
-- +goose Up
CREATE TABLE templates (
    id              BIGSERIAL PRIMARY KEY,
    organisation_id BIGINT NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    html_content    TEXT NOT NULL,
    is_default      BOOLEAN NOT NULL DEFAULT FALSE,
    thumbnail_url   TEXT,
    created_by      BIGINT REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_templates_org ON templates(organisation_id);
-- +goose Down
DROP TABLE IF EXISTS templates;
```

### 016 — whatsapp_logs
```sql
-- +goose Up
CREATE TABLE whatsapp_logs (
    id              BIGSERIAL PRIMARY KEY,
    organisation_id BIGINT NOT NULL REFERENCES organisations(id),
    invoice_id      BIGINT NOT NULL REFERENCES invoices(id),
    recipient_phone TEXT NOT NULL,
    message         TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    error_message   TEXT,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_wa_logs_invoice ON whatsapp_logs(invoice_id);
-- +goose Down
DROP TABLE IF EXISTS whatsapp_logs;
```

### 017 — settings
```sql
-- +goose Up
CREATE TABLE settings (
    id                        BIGSERIAL PRIMARY KEY,
    organisation_id           BIGINT NOT NULL UNIQUE REFERENCES organisations(id) ON DELETE CASCADE,
    business_name             TEXT,
    business_email            TEXT,
    business_phone            TEXT,
    business_address          TEXT,
    logo_url                  TEXT,
    currency                  TEXT NOT NULL DEFAULT 'USD',
    date_format               TEXT NOT NULL DEFAULT 'YYYY-MM-DD',
    invoice_prefix            TEXT NOT NULL DEFAULT 'INV',
    default_due_days          INT NOT NULL DEFAULT 30,
    default_tax_rate          NUMERIC(5,2) NOT NULL DEFAULT 0,
    default_template_id       BIGINT,
    whatsapp_enabled          BOOLEAN NOT NULL DEFAULT FALSE,
    whatsapp_api_key          TEXT,
    whatsapp_message_template TEXT NOT NULL DEFAULT 'Dear {{customer_name}}, please find invoice {{invoice_number}} for {{total}}. Due: {{due_date}}.',
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose Down
DROP TABLE IF EXISTS settings;
```

### 018 — audit_logs
```sql
-- +goose Up
CREATE TABLE audit_logs (
    id              BIGSERIAL PRIMARY KEY,
    organisation_id BIGINT NOT NULL,
    actor_id        BIGINT,
    actor_type      TEXT NOT NULL DEFAULT 'user',
    action          TEXT NOT NULL,
    resource_type   TEXT NOT NULL,
    resource_id     BIGINT,
    old_value       JSONB,
    new_value       JSONB,
    ip_address      INET,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_audit_org      ON audit_logs(organisation_id, created_at DESC);
CREATE INDEX idx_audit_resource ON audit_logs(resource_type, resource_id);
-- +goose Down
DROP TABLE IF EXISTS audit_logs;
```

### 019 — super_audit_logs
```sql
-- +goose Up
CREATE TABLE super_audit_logs (
    id             BIGSERIAL PRIMARY KEY,
    super_admin_id BIGINT REFERENCES super_admins(id),
    action         TEXT NOT NULL,
    target_type    TEXT,
    target_id      BIGINT,
    details        JSONB,
    ip_address     INET,
    user_agent     TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_super_audit ON super_audit_logs(super_admin_id, created_at DESC);
-- +goose Down
DROP TABLE IF EXISTS super_audit_logs;
```

---

## 7. ALL API ROUTES

### SuperAdmin Routes — base prefix: `/superadmin`
> Middleware chain: `RecoveryMiddleware → RequestIDMiddleware → LoggerMiddleware → SecurityHeadersMiddleware → CORSMiddleware → RateLimitMiddleware → SuperAuthMiddleware → SuperRBACMiddleware → SuperAuditMiddleware`

| Method | Route | Role | Description |
|--------|-------|------|-------------|
| POST | `/superadmin/auth/login` | public | SuperAdmin login |
| POST | `/superadmin/auth/refresh` | public | Refresh SuperAdmin access token |
| POST | `/superadmin/auth/logout` | any SA | Revoke SuperAdmin refresh token |
| GET | `/superadmin/organisations` | any SA | List all organisations + usage stats |
| POST | `/superadmin/organisations` | superadmin | Create / onboard new organisation |
| GET | `/superadmin/organisations/:id` | any SA | Get org details |
| PUT | `/superadmin/organisations/:id` | superadmin | Update org details |
| PUT | `/superadmin/organisations/:id/status` | superadmin | Suspend or reactivate org |
| DELETE | `/superadmin/organisations/:id` | superadmin | Delete org + all data (GDPR) |
| PUT | `/superadmin/organisations/:id/plan` | superadmin, finance | Change org plan |
| GET | `/superadmin/organisations/:id/users` | any SA | List users in org |
| GET | `/superadmin/organisations/:id/audit` | any SA | Audit log for org |
| POST | `/superadmin/organisations/:id/impersonate` | superadmin, support | Impersonate org user |
| DELETE | `/superadmin/impersonate/:sessionId` | superadmin, support | End impersonation session |
| GET | `/superadmin/plans` | any SA | List all subscription plans |
| POST | `/superadmin/plans` | superadmin | Create plan |
| PUT | `/superadmin/plans/:id` | superadmin | Update plan |
| DELETE | `/superadmin/plans/:id` | superadmin | Deactivate plan |
| GET | `/superadmin/users` | any SA | List all users across all orgs |
| PUT | `/superadmin/users/:id/status` | superadmin, support | Suspend or reactivate any user |
| POST | `/superadmin/users/:id/reset-password` | superadmin, support | Force password reset |
| GET | `/superadmin/audit` | any SA | Global platform audit log |
| GET | `/superadmin/metrics` | any SA | Platform metrics (MRR, orgs, users, invoices) |
| GET | `/superadmin/config` | any SA | Get global platform config |
| PUT | `/superadmin/config` | superadmin | Update global platform config |
| POST | `/superadmin/maintenance` | superadmin | Toggle maintenance mode |

---

### Organisation Routes — base prefix: `/api/v1`
> Middleware chain: `RecoveryMiddleware → RequestIDMiddleware → LoggerMiddleware → SecurityHeadersMiddleware → CORSMiddleware → RateLimitMiddleware → AuthMiddleware → TenantMiddleware → RBACMiddleware → PlanLimitMiddleware → AuditMiddleware`

#### Auth
| Method | Route | Description |
|--------|-------|-------------|
| POST | `/api/v1/auth/login` | Org user login — returns access + refresh tokens |
| POST | `/api/v1/auth/refresh` | Refresh org access token |
| POST | `/api/v1/auth/logout` | Revoke refresh token |
| GET | `/api/v1/auth/me` | Get current user profile |
| PUT | `/api/v1/auth/me/password` | Change own password |

#### Customers
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/v1/customers` | List customers (paginated, searchable by name/phone/email) |
| POST | `/api/v1/customers` | Create customer |
| GET | `/api/v1/customers/:id` | Get customer by ID |
| PUT | `/api/v1/customers/:id` | Update customer |
| DELETE | `/api/v1/customers/:id` | Soft-delete customer |

#### Services
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/v1/services` | List services (paginated) |
| POST | `/api/v1/services` | Create service |
| GET | `/api/v1/services/:id` | Get service by ID |
| PUT | `/api/v1/services/:id` | Update service |
| DELETE | `/api/v1/services/:id` | Soft-delete service |

#### Invoice Sessions
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/v1/invoice-sessions` | List sessions for this org |
| POST | `/api/v1/invoice-sessions` | Create session for a year |
| GET | `/api/v1/invoice-sessions/:id` | Get session details |

#### Invoices
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/v1/invoices` | List invoices (filter by status, customer, date range, paginated) |
| POST | `/api/v1/invoices` | Create invoice with items |
| GET | `/api/v1/invoices/:id` | Get invoice with items + payment summary |
| PUT | `/api/v1/invoices/:id` | Update invoice (draft status only) |
| DELETE | `/api/v1/invoices/:id` | Cancel invoice |
| PUT | `/api/v1/invoices/:id/status` | Manually update invoice status |
| POST | `/api/v1/invoices/:id/send` | Mark as sent + trigger WhatsApp if enabled |
| POST | `/api/v1/invoices/:id/whatsapp` | Manually send invoice via WhatsApp |
| GET | `/api/v1/invoices/:id/pdf` | Download invoice as PDF |

#### Payments
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/v1/payments` | List payments (paginated) |
| POST | `/api/v1/payments` | Record payment against an invoice |
| GET | `/api/v1/payments/:id` | Get payment by ID |
| DELETE | `/api/v1/payments/:id` | Delete payment record (admin only) |

#### Templates
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/v1/templates` | List templates |
| POST | `/api/v1/templates` | Create template |
| GET | `/api/v1/templates/:id` | Get template |
| PUT | `/api/v1/templates/:id` | Update template |
| DELETE | `/api/v1/templates/:id` | Delete template |
| PUT | `/api/v1/templates/:id/default` | Set as default template |

#### Settings
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/v1/settings` | Get organisation settings |
| PUT | `/api/v1/settings` | Update organisation settings |

#### Audit
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/v1/audit` | List audit log for own org (admin role only) |

---

### System Routes — no auth required
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/health` | Liveness — returns 200 if server is running |
| GET | `/ready` | Readiness — returns 200 if DB is connected |
| GET | `/metrics` | Prometheus metrics endpoint |

---

## 8. MIDDLEWARE — IMPLEMENT ALL OF THESE

### `super_auth.go`
- Extract Bearer token from `Authorization` header
- Verify using `SA_JWT_PUBLIC_KEY` (RS256)
- Load super_admin record from DB, verify `is_active = TRUE`
- Inject into context: `super_admin_id`, `super_admin_role`
- Return 401 if token invalid or expired

### `super_rbac.go`
- Read `super_admin_role` from context
- Compare against required roles for the route
- Return 403 if role is insufficient

### `auth.go`
- Extract Bearer token from `Authorization` header
- Verify using `ORG_JWT_PUBLIC_KEY` (RS256)
- Load user from DB, verify `is_active = TRUE` and org is `active`
- Inject into context: `user_id`, `organisation_id`, `user_role`
- Return 401 if token invalid or expired

### `rbac.go`
- Read `user_role` from context
- Compare against required role for the route
- Return 403 if insufficient

### `tenant.go`
- Load organisation subscription from DB using `organisation_id`
- Verify subscription is `active` or `trialing`
- Inject plan into context for `plan_limit.go`
- Return 403 with `ErrOrgSuspended` if org is suspended

### `plan_limit.go`
- On any POST (create) request check plan limits:
  - Invoices: count current month vs `plan.max_invoices_per_month`
  - Customers: count active vs `plan.max_customers`
  - Users: count active vs `plan.max_users`
- Return 403 with `ErrPlanLimit` if limit exceeded

### `ratelimit.go`
- Auth endpoints: 10 requests/minute per IP
- All other endpoints: 300 requests/minute per authenticated user
- Return 429 with `Retry-After` header

### `security_headers.go`
Set on every response:
```
Strict-Transport-Security : max-age=31536000; includeSubDomains
X-Content-Type-Options    : nosniff
X-Frame-Options           : DENY
Referrer-Policy           : no-referrer
Permissions-Policy        : camera=(), microphone=()
```

### `cors.go`
- Read allowed origins from `ALLOWED_ORIGINS` env var
- Never use wildcard `*`
- Only allow configured origins

### `requestid.go`
- Generate UUID per request, inject as `X-Request-ID` header and into context

### `logger.go`
- Log every request: `method`, `path`, `status`, `latency_ms`, `ip`, `request_id`, `user_id`

### `recovery.go`
- Catch any panic, log stack trace internally, return structured 500 to client

### `error.go`
Map typed errors to HTTP codes:
```
ErrNotFound      → 404    ErrUnauthorised  → 401
ErrForbidden     → 403    ErrValidation    → 422
ErrConflict      → 409    ErrPlanLimit     → 403
ErrOrgSuspended  → 403    ErrAccountLocked → 423
ErrTokenExpired  → 401    ErrMaintenance   → 503
```

---

## 9. CONSTANTS

### `internal/constants/roles.go`
```go
const (
    // Context keys
    CtxUserID               = "user_id"
    CtxOrgID                = "organisation_id"
    CtxUserRole             = "user_role"
    CtxSuperAdminID         = "super_admin_id"
    CtxSuperAdminRole       = "super_admin_role"
    CtxRequestID            = "request_id"
    CtxOrgPlan              = "org_plan"
    CtxIsImpersonating      = "is_impersonating"
    CtxImpersonationSession = "impersonation_session_id"

    // SuperAdmin roles
    RoleSuperAdmin = "superadmin"
    RoleSupport    = "support"
    RoleFinance    = "finance"
    RoleReadonly   = "readonly"

    // Organisation roles
    RoleOrgAdmin   = "admin"
    RoleOrgManager = "manager"
    RoleOrgViewer  = "viewer"
)
```

### `internal/constants/status.go`
```go
const (
    // Invoice
    InvoiceStatusDraft     = "draft"
    InvoiceStatusSent      = "sent"
    InvoiceStatusViewed    = "viewed"
    InvoiceStatusPaid      = "paid"
    InvoiceStatusOverdue   = "overdue"
    InvoiceStatusCancelled = "cancelled"

    // Payment methods
    PaymentCash         = "cash"
    PaymentBankTransfer = "bank_transfer"
    PaymentCard         = "card"
    PaymentCheque       = "cheque"
    PaymentOnline       = "online"
    PaymentOther        = "other"

    // Subscription
    SubActive    = "active"
    SubTrialing  = "trialing"
    SubPastDue   = "past_due"
    SubCancelled = "cancelled"
    SubSuspended = "suspended"

    // Organisation
    OrgActive    = "active"
    OrgSuspended = "suspended"
    OrgCancelled = "cancelled"
    OrgTrial     = "trial"

    // WhatsApp logs
    WAPending   = "pending"
    WASent      = "sent"
    WADelivered = "delivered"
    WAFailed    = "failed"
)
```

### `internal/constants/errors.go`
```go
var (
    ErrNotFound       = errors.New("resource not found")
    ErrUnauthorised   = errors.New("unauthorised")
    ErrForbidden      = errors.New("forbidden")
    ErrValidation     = errors.New("validation failed")
    ErrConflict       = errors.New("resource already exists")
    ErrPlanLimit      = errors.New("plan limit reached — upgrade your plan")
    ErrOrgSuspended   = errors.New("organisation is suspended")
    ErrAccountLocked  = errors.New("account locked — too many failed attempts")
    ErrTokenExpired   = errors.New("token has expired")
    ErrTokenInvalid   = errors.New("token is invalid")
    ErrMaintenance    = errors.New("platform is under maintenance")
    ErrInternalServer = errors.New("internal server error")
)
```

---

## 10. RESPONSE FORMAT

### Success
```json
{
    "success": true,
    "message": "Customers retrieved successfully",
    "data": {},
    "meta": { "page": 1, "per_page": 20, "total": 150, "total_pages": 8 },
    "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Error
```json
{
    "success": false,
    "message": "Validation failed",
    "errors": [
        { "field": "email", "message": "must be a valid email address" }
    ],
    "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

Implement helpers in `internal/utils/response.go`: `SuccessResponse()`, `ErrorResponse()`, `ValidationErrorResponse()`.

---

## 11. BUSINESS LOGIC — IMPLEMENT EXACTLY AS DESCRIBED

### Logic 1 — Atomic Invoice Number Generation
```
When creating an invoice:
1. Get current year
2. Find invoice_session WHERE (organisation_id, year, prefix)
   If none exists → create one inside same transaction
3. Atomically increment (sqlc query):
   UPDATE invoice_sessions
   SET current_sequence = current_sequence + 1
   WHERE organisation_id = $1 AND year = $2 AND prefix = $3
   RETURNING current_sequence, prefix, year;
4. Format: {prefix}-{year}-{sequence zero-padded to 4 digits}
   Example: INV-2025-0001

NEVER use SELECT then INSERT for number generation.
This causes duplicate numbers under concurrent load.
```

### Logic 2 — Service Auto-Fill on Invoice Item
```
If service_id is provided in an invoice item:
  1. Fetch service record
  2. COPY into invoice_item:
       description ← service.description
       unit_price  ← service.unit_price
       tax_rate    ← service.tax_rate
  3. Values are COPIED not referenced (historical accuracy)
  4. User may override before saving
```

### Logic 3 — Auto Customer Creation
```
On POST /api/v1/invoices:
  If customer_id is provided → use it directly
  If customer object provided (no customer_id):
    Search: SELECT id FROM customers
            WHERE organisation_id = $1
            AND (email = $2 OR phone = $3) LIMIT 1
    If found → use existing id
    If not found → INSERT new customer in SAME transaction
  Always return resolved customer_id in response
```

### Logic 4 — Invoice Total Calculation
```
Per item:
  tax_amount = ROUND(quantity × unit_price × (tax_rate / 100), 2)
  line_total  = ROUND((quantity × unit_price) + tax_amount, 2)

Invoice:
  subtotal        = SUM(quantity × unit_price)
  tax_amount      = SUM(item.tax_amount)
  discount_amount = applied discount (0 if none)
  total           = subtotal + tax_amount - discount_amount

All calculations in service layer. Store all values in DB.
```

### Logic 5 — Auto Invoice Status on Payment
```
After recording a payment:
  total_paid = SELECT SUM(amount) FROM payments WHERE invoice_id = $1

  IF total_paid >= invoice.total:
    UPDATE invoices SET status = 'paid', paid_at = NOW()

  IF 0 < total_paid < invoice.total:
    Status remains 'sent' (partial payment noted)
```

### Logic 6 — Async WhatsApp Sending
```
1. Check whatsapp_enabled = TRUE in settings → else return 422
2. INSERT into whatsapp_logs with status = 'pending'
3. Return HTTP 202 Accepted immediately
4. Goroutine:
   a. Call WhatsApp API
   b. On success: UPDATE whatsapp_logs SET status='sent', sent_at=NOW()
   c. On failure: UPDATE whatsapp_logs SET status='failed', error_message='...'
```

### Logic 7 — Plan Limit Checks
```
POST /api/v1/invoices:
  IF monthly invoice count >= plan.max_invoices_per_month → 403

POST /api/v1/customers:
  IF active customer count >= plan.max_customers → 403

POST /api/v1/users:
  IF active user count >= plan.max_users → 403
```

### Logic 8 — Impersonation Flow
```
POST /superadmin/organisations/:id/impersonate
Body: { "target_user_id": 5, "reason": "Support ticket #1234" }

1. Verify super_admin role is 'superadmin' or 'support'
2. Verify org status = 'active'
3. Verify target user is_active = TRUE
4. Reject if reason is empty string
5. INSERT into impersonation_sessions
6. Issue Org JWT (1 hour expiry) with extra claims:
     "impersonated_by": super_admin_id
     "impersonation_session_id": session_id
7. All actions with this token log actor_type = 'superadmin'
```

---

## 12. SECURITY REQUIREMENTS

```
JWT:
  - RS256 asymmetric — NEVER HS256
  - Two separate key pairs (org + superadmin)
  - Access token expiry  : 15 minutes
  - Refresh token expiry : 7 days (org), 24 hours (superadmin)
  - Store refresh tokens as SHA-256 hash in DB, never plaintext

Passwords:
  - bcrypt cost 12 only — no MD5, SHA, or plain text
  - Minimum 8 characters
  - Lock account after 5 failed attempts for 15 minutes

Rate Limiting:
  - Login endpoints : 10 req/min per IP
  - API endpoints   : 300 req/min per user
  - Return 429 with Retry-After header

CORS:
  - Never wildcard *
  - Whitelist from ALLOWED_ORIGINS env only

Input Validation:
  - Max body: 1MB JSON, 2MB file uploads
  - Validate all fields with go-playground/validator
  - File uploads: .png .jpg .webp only
  - Never return raw DB errors to client

SuperAdmin:
  - Separate JWT key pair from org users
  - IP allowlist via SA_IP_ALLOWLIST env var
  - All mutations logged to super_audit_logs automatically
  - Impersonation requires non-empty reason

Database:
  - ALL org-scoped queries MUST include WHERE organisation_id = $n
  - Never SELECT * — list columns explicitly in sqlc queries
  - Parameterised queries only — never string concatenation
```

---

## 13. SQLC CONFIGURATION — `sqlc.yaml`

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/db/queries"
    schema: "internal/db/migrations"
    gen:
      go:
        package: "db"
        out: "internal/db/sqlc"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        null_style: "sql_null_types"
```

---

## 14. ENVIRONMENT FILE — `.env.example`

```env
# Server
SERVER_PORT=8080
ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# Database (Neon / PostgreSQL)
DATABASE_URL=postgres://user:password@localhost:5432/invoice_db?sslmode=disable
DB_MIN_CONNS=5
DB_MAX_CONNS=25
DB_MAX_CONN_LIFETIME=1h
DB_MAX_CONN_IDLE_TIME=30m

# JWT — Organisation Users
ORG_JWT_PRIVATE_KEY_PATH=./keys/org_private.pem
ORG_JWT_PUBLIC_KEY_PATH=./keys/org_public.pem
ORG_ACCESS_TOKEN_EXPIRY=15m
ORG_REFRESH_TOKEN_EXPIRY=168h

# JWT — SuperAdmin
SA_JWT_PRIVATE_KEY_PATH=./keys/sa_private.pem
SA_JWT_PUBLIC_KEY_PATH=./keys/sa_public.pem
SA_ACCESS_TOKEN_EXPIRY=15m
SA_REFRESH_TOKEN_EXPIRY=24h

# SuperAdmin Security
SA_IP_ALLOWLIST=127.0.0.1

# Rate Limiting
RATE_LIMIT_AUTH_RPM=10
RATE_LIMIT_API_RPM=300

# File Uploads
UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE_MB=2

# WhatsApp
WHATSAPP_API_URL=
WHATSAPP_API_KEY=

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## 15. GRACEFUL SHUTDOWN — `cmd/server/main.go`

```go
srv := &http.Server{
    Addr:         ":" + cfg.ServerPort,
    Handler:      router,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}

go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        logger.Fatal("Server failed to start", zap.Error(err))
    }
}()

logger.Info("Server started", zap.String("port", cfg.ServerPort))

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

logger.Info("Shutdown signal received — draining connections")
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
    logger.Fatal("Forced shutdown", zap.Error(err))
}
dbPool.Close()
logger.Info("Server exited cleanly")
```

---

## 16. .gitignore

```gitignore
# Environment
.env
.env.*
!.env.example

# JWT keys — never commit
keys/

# Build output
bin/

# Uploads
uploads/logos/*
uploads/invoices/*
!uploads/logos/.gitkeep
!uploads/invoices/.gitkeep

# Go
*.exe
*.test
coverage.out

# IDE
.idea/
.vscode/
*.swp
```

---

## 17. FILE GENERATION CHECKLIST

Generate all files in this exact order:

```
 1.  go.mod                                      — Section 2
 2.  .env.example                                — Section 14
 3.  .gitignore                                  — Section 16
 4.  sqlc.yaml                                   — Section 13
 5.  README.md                                   — all go run / goose / sqlc commands from Section 3
 6.  internal/constants/roles.go                 — Section 9
 7.  internal/constants/status.go                — Section 9
 8.  internal/constants/errors.go                — Section 9
 9.  internal/config/config.go                   — Viper struct for all .env values
10.  internal/utils/response.go                  — SuccessResponse, ErrorResponse helpers
11.  internal/utils/jwt.go                       — RS256 issue / verify / refresh (both key pairs)
12.  internal/utils/hash.go                      — bcrypt hash + compare (cost 12)
13.  internal/utils/invoice_number.go            — atomic generator (Logic 1)
14.  internal/utils/pagination.go                — cursor + offset helpers
15.  internal/utils/sanitize.go                  — input sanitisation
16.  internal/db/migrations/001 to 019           — all migration SQL files from Section 6
17.  internal/db/queries/*.sql                   — sqlc query files (one per table)
18.  internal/db/sqlc/                           — run: sqlc generate
19.  internal/middleware/recovery.go
20.  internal/middleware/requestid.go
21.  internal/middleware/logger.go
22.  internal/middleware/security_headers.go
23.  internal/middleware/cors.go
24.  internal/middleware/ratelimit.go
25.  internal/middleware/super_auth.go
26.  internal/middleware/super_rbac.go
27.  internal/middleware/auth.go
28.  internal/middleware/rbac.go
29.  internal/middleware/tenant.go
30.  internal/middleware/plan_limit.go
31.  internal/middleware/error.go
32.  internal/modules/superadmin/auth/           — handler, service, routes, validation
33.  internal/modules/superadmin/organisations/  — handler, service, routes
34.  internal/modules/superadmin/plans/          — handler, service, routes
35.  internal/modules/superadmin/users/          — handler, service, routes
36.  internal/modules/superadmin/impersonation/  — handler, service, routes
37.  internal/modules/superadmin/metrics/        — handler, routes
38.  internal/modules/superadmin/config/         — handler, service, routes
39.  internal/modules/auth/                      — handler, service, routes, validation
40.  internal/modules/customers/                 — handler, service, routes, validation
41.  internal/modules/services/                  — handler, service, routes, validation
42.  internal/modules/invoice_sessions/          — handler, service, routes
43.  internal/modules/invoices/                  — handler, service, routes, validation (Logics 1–4)
44.  internal/modules/invoice_items/             — handler, service
45.  internal/modules/payments/                  — handler, service, routes, validation (Logic 5)
46.  internal/modules/templates/                 — handler, service, routes
47.  internal/modules/whatsapp/                  — handler, service, routes (Logic 6)
48.  internal/modules/settings/                  — handler, service, routes
49.  internal/modules/audit/                     — handler, service, routes
50.  cmd/server/main.go                          — server init, all routes, graceful shutdown
51.  uploads/logos/.gitkeep
52.  uploads/invoices/.gitkeep
53.  tests/unit/.gitkeep
54.  tests/integration/.gitkeep
```

---

*End of specification. Generate the complete project now. Do not skip any file.*