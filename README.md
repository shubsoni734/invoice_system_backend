# Invoice Backend

Multi-tenant SaaS invoice management platform built with Go 1.22+.

## Tech Stack

- **Language**: Go 1.22+
- **Framework**: Gin
- **Database**: PostgreSQL 16 (Neon)
- **Migrations**: Goose
- **Query Layer**: sqlc (type-safe generated DB code)
- **Config**: Viper
- **Logging**: Zap (structured JSON)
- **JWT**: golang-jwt/jwt (RS256, asymmetric)
- **Password**: bcrypt (cost 12)

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go                        # Application entry point
│
├── internal/
│   ├── app/
│   │   └── routes.go                      # Top-level route registration
│   │
│   ├── config/
│   │   └── config.go                      # Environment config loading (Viper)
│   │
│   ├── domain/                            # Business domains (DDD)
│   │   └── superadmin/
│   │       ├── auth/
│   │       │   ├── handler.go             # HTTP handlers
│   │       │   ├── routes.go              # Route registration
│   │       │   ├── queries.sql            # Raw SQL queries (sqlc source)
│   │       │   └── sqlc/                  # Generated type-safe DB code
│   │       │       ├── db.go
│   │       │       ├── models.go
│   │       │       └── queries.sql.go
│   │       └── organisations/
│   │           ├── handler.go
│   │           ├── routes.go
│   │           ├── queries.sql
│   │           └── sqlc/
│   │               ├── db.go
│   │               ├── models.go
│   │               └── queries.sql.go
│   │
│   ├── pkg/                               # Shared internal packages
│   │   ├── db/
│   │   │   └── migrations/               # 19 Goose SQL migration files
│   │   ├── middleware/                   # HTTP middleware (auth, cors, rate limit, etc.)
│   │   ├── response/                     # Standardized JSON response helpers
│   │   └── utils/                        # JWT, hash, pagination, sanitize, invoice number
│   │
│   └── shared/
│       └── constants/                    # Roles, statuses, error types, context keys
│
├── keys/                                 # RSA key pairs (gitignored)
├── uploads/                              # File upload storage
├── sqlc.yaml                             # sqlc configuration
├── .env.development
├── .env.production
├── go.mod
├── go.sum
└── Makefile
```

## Prerequisites

- Go 1.22+
- PostgreSQL 16 (or Neon cloud)
- [goose](https://github.com/pressly/goose) for migrations
- [sqlc](https://sqlc.dev) for query generation

## First-Time Setup

### 1. Install tools
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### 2. Configure environment
```bash
cp .env.example .env
# Edit .env and set DATABASE_URL, ORG_JWT_SECRET, SA_JWT_SECRET
# Add BREVO_API_KEY and BREVO_SENDER_EMAIL when you are ready to send email
```

Email is sent through the Brevo transactional API. The server starts without those keys. Welcome, password-reset, invoice, and test messages fail with a configuration error until `BREVO_API_KEY` and `BREVO_SENDER_EMAIL` are set. In production, set `SA_BOOTSTRAP_SECRET` and send it as `X-Bootstrap-Secret` to create the first superadmin.

### 4. Run migrations
```bash
goose -dir internal/pkg/db/migrations postgres "$DATABASE_URL" up
```

### 5. Regenerate sqlc (after editing any queries.sql)
```bash
sqlc generate
```

### 6. Start the server
```bash
go run ./cmd/api/main.go
# or
make run
```

## Makefile Commands

```bash
make run            # Run the server
make build          # Build binary to bin/
make tidy           # go mod tidy
make test           # Run unit tests
make sqlc           # Regenerate sqlc code
make migrate-up     # Run all pending migrations
make migrate-down   # Roll back last migration
make migrate-status # Show migration status
```

## API Endpoints

Server runs on `http://localhost:8080`

### System
- `GET /health` — Liveness probe
- `GET /ready` — Readiness probe (checks DB)

### SuperAdmin Auth (public)
- `POST /superadmin/auth/create` — Create superadmin account
- `POST /superadmin/auth/login` — Login, returns access + refresh tokens

### SuperAdmin Auth (Bearer token required)
- `GET /superadmin/auth/me` — Get current superadmin profile
- `POST /superadmin/auth/logout` — Revoke all refresh tokens

### SuperAdmin Organisations (Bearer token required)
- `POST /superadmin/organisations` — Create organisation
- `GET /superadmin/organisations` — List organisations (paginated)
- `GET /superadmin/organisations/:id` — Get organisation + active subscription
- `POST /superadmin/organisations/:id/subscription` — Apply/change subscription plan

## Database Tables (19 Total)

| # | Table |
|---|-------|
| 1 | super_admins |
| 2 | plans |
| 3 | organisations |
| 4 | organisation_subscriptions |
| 5 | users |
| 6 | refresh_tokens |
| 7 | super_refresh_tokens |
| 8 | impersonation_sessions |
| 9 | customers |
| 10 | services |
| 11 | invoice_sessions |
| 12 | invoices |
| 13 | invoice_items |
| 14 | payments |
| 15 | templates |
| 16 | whatsapp_logs |
| 17 | settings |
| 18 | audit_logs |
| 19 | super_audit_logs |

All tables use UUID primary keys (`uuid_generate_v4()`).

## JWT Authentication

Two separate JWT managers are created at startup — one for org users, one for superadmin. Both use HS256 (HMAC-SHA256) with a secret string from env.

| Manager | Secret env var | Used in |
|---------|---------------|---------|
| `orgJWT` | `ORG_JWT_SECRET` | `middleware.Auth` — protects `/api/v1/*` routes |
| `superJWT` | `SA_JWT_SECRET` | `middleware.SuperAuth` — protects `/superadmin/*` routes |

### Token flow

```
POST /superadmin/auth/login
  → handler.go: GenerateToken (15m) + GenerateRefreshToken (24h)
  → refresh token hash stored in super_refresh_tokens table

GET /superadmin/auth/me  (Bearer <access_token>)
  → middleware/super_auth.go: VerifyToken → sets ctx super_admin_id
  → handler.go: reads ctx super_admin_id, queries DB

POST /superadmin/auth/logout
  → handler.go: RevokeAllSuperRefreshTokens for that super_admin_id
```

### Claims structure

```go
type Claims struct {
    UserID               string  // org user ID
    OrgID                string  // organisation ID
    Role                 string  // user/superadmin role
    SuperAdminID         string  // set for superadmin tokens
    ImpersonatedBy       string  // set during impersonation
    ImpersonationSession string
    jwt.RegisteredClaims        // exp, iat
}
```



## Security

- JWT HS256 — separate secrets for org users and superadmin (`ORG_JWT_SECRET`, `SA_JWT_SECRET`)
- Bcrypt password hashing (cost 12)
- Account lockout after 5 failed attempts (15 min cooldown)
- Rate limiting: 10 req/min (auth), 300 req/min (API)
- IP allowlist for SuperAdmin panel access
- Security headers: HSTS, X-Frame-Options, X-Content-Type-Options, CSP
- CORS with strict origin whitelist

## License

Proprietary — All rights reserved
