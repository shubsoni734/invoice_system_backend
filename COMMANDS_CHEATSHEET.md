# Commands Cheatsheet

## 🚀 Quick Setup

```powershell
# 1. Setup
cp .env.example .env
# Edit .env with DATABASE_URL

# 2. Generate keys
go run generate_keys.go

# 3. Run migrations
.\scripts\run_migrations.ps1

# 4. Create SuperAdmin
.\scripts\create_superadmin.ps1

# 5. Start server
go run cmd/server/main.go
```

## 📊 Database Migrations

```bash
# Run all migrations (Recommended)
.\scripts\run_migrations.ps1              # Windows
./scripts/run_migrations.sh               # Linux/Mac

# Manual migration commands
goose -dir internal/db/migrations postgres "$DATABASE_URL" up      # Apply all
goose -dir internal/db/migrations postgres "$DATABASE_URL" down    # Rollback one
goose -dir internal/db/migrations postgres "$DATABASE_URL" status  # Check status
goose -dir internal/db/migrations postgres "$DATABASE_URL" reset   # Reset all (⚠️ DESTRUCTIVE)
```

## 👤 SuperAdmin

```bash
# Create SuperAdmin
.\scripts\create_superadmin.ps1           # Windows
./scripts/create_superadmin.sh            # Linux/Mac

# With custom credentials
$env:SUPERADMIN_EMAIL="admin@example.com"
$env:SUPERADMIN_PASSWORD="YourPassword123!"
.\scripts\create_superadmin.ps1
```

## 🔧 Development

```bash
# Install dependencies
go mod download

# Run server
go run cmd/server/main.go

# Build binary
go build -o invoicepro cmd/server/main.go

# Run binary
./invoicepro

# Run tests
go test ./... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 🧪 Testing

```bash
# Health check
curl http://localhost:8080/health

# Ready check (tests DB)
curl http://localhost:8080/ready

# SuperAdmin login
curl -X POST http://localhost:8080/superadmin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"superadmin@invoicepro.com","password":"SuperAdmin@123"}'

# Organisation login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'
```

## 📦 Postman

```powershell
# Generate collection
.\scripts\generate_postman_collection.ps1

# Import to Postman
# File: postman/InvoicePro_API_Collection.json
```

## 🔍 Debugging

```bash
# Check if port is in use
netstat -ano | findstr :8080              # Windows
lsof -ti:8080                             # Linux/Mac

# Kill process on port
taskkill /PID <PID> /F                    # Windows
kill -9 $(lsof -ti:8080)                  # Linux/Mac

# Check database connection
psql "$DATABASE_URL"

# View logs
# Logs are output to console by default
```

## 🗄️ Database

```bash
# Connect to database
psql "$DATABASE_URL"

# List tables
\dt

# Describe table
\d table_name

# View table data
SELECT * FROM super_admins;

# Check migration status
SELECT * FROM goose_db_version;
```

## 📝 Environment Variables

```bash
# Windows PowerShell
$env:DATABASE_URL="postgresql://..."
$env:SERVER_PORT="8080"

# Linux/Mac
export DATABASE_URL="postgresql://..."
export SERVER_PORT="8080"

# Load from .env
Get-Content .env | ForEach-Object {       # Windows
    if ($_ -match '^([^=]+)=(.*)$') {
        [Environment]::SetEnvironmentVariable($matches[1], $matches[2], "Process")
    }
}

export $(cat .env | xargs)                # Linux/Mac
```

## 🛠️ Tools Installation

```bash
# Goose (migrations)
go install github.com/pressly/goose/v3/cmd/goose@latest

# SQLC (query generator)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Air (hot reload)
go install github.com/cosmtrek/air@latest
```

## 🔐 JWT Keys

```bash
# Generate keys
go run generate_keys.go

# Manual generation
mkdir keys
openssl genrsa -out keys/org_private.pem 2048
openssl rsa -in keys/org_private.pem -pubout -out keys/org_public.pem
openssl genrsa -out keys/sa_private.pem 2048
openssl rsa -in keys/sa_private.pem -pubout -out keys/sa_public.pem
```

## 📚 Documentation

```bash
# View documentation
cat START_HERE.md                         # Quick start
cat DATABASE_SETUP.md                     # Database guide
cat API_ENDPOINTS.md                      # API reference
cat SUPERADMIN_SETUP.md                   # SuperAdmin guide
```

## 🚨 Emergency Commands

```bash
# Reset database (⚠️ DESTRUCTIVE)
goose -dir internal/db/migrations postgres "$DATABASE_URL" reset
goose -dir internal/db/migrations postgres "$DATABASE_URL" up

# Force kill server
taskkill /F /IM go.exe                    # Windows
pkill -9 go                               # Linux/Mac

# Clear Go cache
go clean -cache -modcache -testcache

# Reinstall dependencies
rm go.sum
go mod download
```

## 📊 Useful Queries

```sql
-- Check all tables
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public';

-- Count records
SELECT 
    schemaname,
    tablename,
    n_live_tup as row_count
FROM pg_stat_user_tables
ORDER BY n_live_tup DESC;

-- Check SuperAdmin
SELECT id, email, role, is_active, created_at 
FROM super_admins;

-- Check organisations
SELECT id, name, slug, status, created_at 
FROM organisations;

-- Check users
SELECT u.id, u.name, u.email, u.role, o.name as org_name
FROM users u
JOIN organisations o ON u.organisation_id = o.id;
```

---

**Quick Reference:** Keep this file handy for common commands!
