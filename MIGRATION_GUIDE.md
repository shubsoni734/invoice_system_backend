# Database Migration Guide

## Quick Start (3 Steps)

### Step 1: Install Goose
```powershell
# Windows PowerShell
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### Step 2: Set Database URL
```powershell
# Windows PowerShell
$env:DATABASE_URL="postgresql://neondb_owner:npg_GRkOWz9Hf6rI@ep-mute-mud-adi7b3r8-pooler.c-2.us-east-1.aws.neon.tech/neondb?sslmode=require"
```

Or add to your `.env` file:
```env
DATABASE_URL=postgresql://neondb_owner:npg_GRkOWz9Hf6rI@ep-mute-mud-adi7b3r8-pooler.c-2.us-east-1.aws.neon.tech/neondb?sslmode=require
```

### Step 3: Run Migrations
```powershell
# Windows PowerShell
goose -dir internal/db/migrations postgres $env:DATABASE_URL up
```

**Done!** All 19 tables will be created.

---

## Detailed Instructions

### Installation

#### Windows
```powershell
# Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Verify installation
goose -version
```

#### Linux/Mac
```bash
# Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Verify installation
goose -version
```

### Configuration

#### Option 1: Environment Variable (Recommended)

**Windows PowerShell:**
```powershell
$env:DATABASE_URL="your_database_url_here"
```

**Linux/Mac Bash:**
```bash
export DATABASE_URL="your_database_url_here"
```

#### Option 2: .env File

Add to `.env`:
```env
DATABASE_URL=postgresql://user:password@host:port/database?sslmode=require
```

Then load it:
```powershell
# Windows PowerShell
Get-Content .env | ForEach-Object {
    if ($_ -match '^([^=]+)=(.*)$') {
        [Environment]::SetEnvironmentVariable($matches[1], $matches[2])
    }
}
```

### Running Migrations

#### Run All Migrations (Up)
```powershell
# Windows PowerShell
goose -dir internal/db/migrations postgres $env:DATABASE_URL up
```

```bash
# Linux/Mac
goose -dir internal/db/migrations postgres "$DATABASE_URL" up
```

**Output:**
```
OK    001_create_super_admins.sql (123.45ms)
OK    002_create_plans.sql (89.12ms)
OK    003_create_organisations.sql (95.67ms)
...
OK    019_create_super_audit_logs.sql (78.34ms)
goose: successfully migrated database to version: 19
```

#### Check Migration Status
```powershell
goose -dir internal/db/migrations postgres $env:DATABASE_URL status
```

**Output:**
```
Applied At                  Migration
=======================================
2024-03-10 10:30:45 UTC  -- 001_create_super_admins.sql
2024-03-10 10:30:45 UTC  -- 002_create_plans.sql
...
Pending                  -- (none)
```

#### Rollback Last Migration
```powershell
goose -dir internal/db/migrations postgres $env:DATABASE_URL down
```

#### Rollback All Migrations
```powershell
goose -dir internal/db/migrations postgres $env:DATABASE_URL reset
```

#### Migrate to Specific Version
```powershell
# Migrate to version 5
goose -dir internal/db/migrations postgres $env:DATABASE_URL up-to 5
```

#### Redo Last Migration
```powershell
# Rollback and re-apply last migration
goose -dir internal/db/migrations postgres $env:DATABASE_URL redo
```

---

## Migration Files

Your project has 19 migration files:

```
internal/db/migrations/
├── 001_create_super_admins.sql
├── 002_create_plans.sql
├── 003_create_organisations.sql
├── 004_create_organisation_subscriptions.sql
├── 005_create_users.sql
├── 006_create_refresh_tokens.sql
├── 007_create_super_refresh_tokens.sql
├── 008_create_impersonation_sessions.sql
├── 009_create_customers.sql
├── 010_create_services.sql
├── 011_create_invoice_sessions.sql
├── 012_create_invoices.sql
├── 013_create_invoice_items.sql
├── 014_create_payments.sql
├── 015_create_templates.sql
├── 016_create_whatsapp_logs.sql
├── 017_create_settings.sql
├── 018_create_audit_logs.sql
└── 019_create_super_audit_logs.sql
```

### Tables Created

1. **super_admins** - SuperAdmin users
2. **plans** - Subscription plans
3. **organisations** - Tenant organisations
4. **organisation_subscriptions** - Org subscriptions
5. **users** - Organisation users
6. **refresh_tokens** - User refresh tokens
7. **super_refresh_tokens** - SuperAdmin refresh tokens
8. **impersonation_sessions** - SuperAdmin impersonation
9. **customers** - Customer records
10. **services** - Service catalog
11. **invoice_sessions** - Invoice numbering
12. **invoices** - Invoice records
13. **invoice_items** - Invoice line items
14. **payments** - Payment records
15. **templates** - Invoice templates
16. **whatsapp_logs** - WhatsApp message logs
17. **settings** - Organisation settings
18. **audit_logs** - Organisation audit logs
19. **super_audit_logs** - SuperAdmin audit logs

---

## Creating New Migrations

### Create New Migration File
```powershell
goose -dir internal/db/migrations create add_new_feature sql
```

This creates:
```
internal/db/migrations/020_add_new_feature.sql
```

### Migration File Format

```sql
-- +goose Up
-- SQL in this section is executed when migrating up
CREATE TABLE example (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
-- SQL in this section is executed when migrating down
DROP TABLE IF EXISTS example;
```

**Important:**
- Always include `-- +goose Up` and `-- +goose Down` comments
- Down migrations should reverse Up migrations
- Test both up and down migrations

---

## Troubleshooting

### Error: "goose: command not found"

**Solution:**
```powershell
# Reinstall goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Add Go bin to PATH
$env:PATH += ";$env:USERPROFILE\go\bin"
```

### Error: "dial tcp: lookup failed"

**Solution:**
Check your DATABASE_URL:
```powershell
# Print current DATABASE_URL
echo $env:DATABASE_URL

# Set correct URL
$env:DATABASE_URL="postgresql://user:password@host:port/database?sslmode=require"
```

### Error: "pq: SSL is not enabled"

**Solution:**
Add `sslmode=disable` for local databases:
```
postgresql://user:password@localhost:5432/database?sslmode=disable
```

Or `sslmode=require` for cloud databases:
```
postgresql://user:password@host:port/database?sslmode=require
```

### Error: "no such table: goose_db_version"

**Solution:**
This is normal for first run. Goose will create this table automatically.

### Error: "migration failed"

**Solution:**
1. Check migration SQL syntax
2. Check database permissions
3. Check if table already exists
4. View detailed error in output

### Reset Database (⚠️ DESTRUCTIVE)

```powershell
# This will drop all tables and re-run migrations
goose -dir internal/db/migrations postgres $env:DATABASE_URL reset
goose -dir internal/db/migrations postgres $env:DATABASE_URL up
```

---

## Database Connection Strings

### Local PostgreSQL
```
postgresql://postgres:password@localhost:5432/invoicepro?sslmode=disable
```

### Neon (Cloud)
```
postgresql://user:password@host.neon.tech/database?sslmode=require
```

### Heroku
```
postgresql://user:password@host.compute.amazonaws.com:5432/database?sslmode=require
```

### Docker
```
postgresql://postgres:password@localhost:5432/invoicepro?sslmode=disable
```

---

## Complete Workflow

### First Time Setup

```powershell
# 1. Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# 2. Set database URL
$env:DATABASE_URL="your_database_url"

# 3. Run migrations
goose -dir internal/db/migrations postgres $env:DATABASE_URL up

# 4. Verify
goose -dir internal/db/migrations postgres $env:DATABASE_URL status
```

### Development Workflow

```powershell
# Create new migration
goose -dir internal/db/migrations create add_feature sql

# Edit the migration file
# Add SQL to Up and Down sections

# Test migration up
goose -dir internal/db/migrations postgres $env:DATABASE_URL up

# Test migration down
goose -dir internal/db/migrations postgres $env:DATABASE_URL down

# Re-apply
goose -dir internal/db/migrations postgres $env:DATABASE_URL up
```

### Production Deployment

```powershell
# 1. Backup database first!
# 2. Check current version
goose -dir internal/db/migrations postgres $env:DATABASE_URL status

# 3. Run migrations
goose -dir internal/db/migrations postgres $env:DATABASE_URL up

# 4. Verify
goose -dir internal/db/migrations postgres $env:DATABASE_URL status
```

---

## Best Practices

1. **Always backup before migrations** in production
2. **Test migrations** in development first
3. **Write reversible migrations** (proper Down sections)
4. **One change per migration** for easier rollback
5. **Never edit applied migrations** - create new ones
6. **Use transactions** when possible
7. **Test rollback** (down) before deploying

---

## Quick Commands Reference

```powershell
# Install
go install github.com/pressly/goose/v3/cmd/goose@latest

# Set DB URL
$env:DATABASE_URL="postgresql://..."

# Run all migrations
goose -dir internal/db/migrations postgres $env:DATABASE_URL up

# Check status
goose -dir internal/db/migrations postgres $env:DATABASE_URL status

# Rollback last
goose -dir internal/db/migrations postgres $env:DATABASE_URL down

# Reset all (⚠️ DESTRUCTIVE)
goose -dir internal/db/migrations postgres $env:DATABASE_URL reset

# Create new migration
goose -dir internal/db/migrations create migration_name sql

# Redo last migration
goose -dir internal/db/migrations postgres $env:DATABASE_URL redo

# Migrate to version
goose -dir internal/db/migrations postgres $env:DATABASE_URL up-to 5
```

---

## Next Steps

After running migrations:

1. ✅ Verify tables created: Check database
2. ✅ Create SuperAdmin: Run `.\scripts\create_superadmin.ps1`
3. ✅ Start server: `go run cmd/server/main.go`
4. ✅ Test API: Import Postman collection

---

**Need Help?**

- Check `START_HERE.md` for complete setup
- See `API_ENDPOINTS.md` for API documentation
- Review `COMPLETE_SETUP_GUIDE.md` for detailed instructions
