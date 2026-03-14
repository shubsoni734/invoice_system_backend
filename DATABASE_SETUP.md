# Database Setup Guide

## Quick Setup (3 Steps)

### Step 1: Configure Database URL

Edit `.env` file and add your Neon database URL:

```env
DATABASE_URL=postgresql://username:password@host:port/database?sslmode=require
```

**Example Neon URL:**
```env
DATABASE_URL=postgresql://neondb_owner:npg_xxxxx@ep-xxxxx.us-east-1.aws.neon.tech/neondb?sslmode=require
```

### Step 2: Run Migrations

```powershell
# Windows
.\scripts\run_migrations.ps1

# Linux/Mac
chmod +x ./scripts/run_migrations.sh
./scripts/run_migrations.sh
```

### Step 3: Verify Tables

The script will create 19 tables:
- ✅ super_admins
- ✅ plans
- ✅ organisations
- ✅ organisation_subscriptions
- ✅ users
- ✅ refresh_tokens
- ✅ super_refresh_tokens
- ✅ impersonation_sessions
- ✅ customers
- ✅ services
- ✅ invoice_sessions
- ✅ invoices
- ✅ invoice_items
- ✅ payments
- ✅ templates
- ✅ whatsapp_logs
- ✅ settings
- ✅ audit_logs
- ✅ super_audit_logs

## Manual Migration Commands

If you prefer to run migrations manually:

### Install Goose (Migration Tool)

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### Run All Migrations

```bash
# Windows PowerShell
goose -dir internal/db/migrations postgres $env:DATABASE_URL up

# Linux/Mac/Git Bash
goose -dir internal/db/migrations postgres "$DATABASE_URL" up
```

### Check Migration Status

```bash
# Windows PowerShell
goose -dir internal/db/migrations postgres $env:DATABASE_URL status

# Linux/Mac/Git Bash
goose -dir internal/db/migrations postgres "$DATABASE_URL" status
```

### Rollback Last Migration

```bash
# Windows PowerShell
goose -dir internal/db/migrations postgres $env:DATABASE_URL down

# Linux/Mac/Git Bash
goose -dir internal/db/migrations postgres "$DATABASE_URL" down
```

### Reset Database (⚠️ DESTRUCTIVE)

```bash
# This will drop all tables and re-run migrations
# Windows PowerShell
goose -dir internal/db/migrations postgres $env:DATABASE_URL reset
goose -dir internal/db/migrations postgres $env:DATABASE_URL up

# Linux/Mac/Git Bash
goose -dir internal/db/migrations postgres "$DATABASE_URL" reset
goose -dir internal/db/migrations postgres "$DATABASE_URL" up
```

## Neon Database Setup

### 1. Get Your Neon Database URL

1. Go to https://console.neon.tech
2. Select your project
3. Click "Connection Details"
4. Copy the connection string
5. It should look like:
   ```
   postgresql://neondb_owner:npg_xxxxx@ep-xxxxx.us-east-1.aws.neon.tech/neondb?sslmode=require
   ```

### 2. Update .env File

```env
DATABASE_URL=your_neon_connection_string_here
```

### 3. Test Connection

```bash
# Using psql (if installed)
psql "your_neon_connection_string"

# Or use the migration script which will test the connection
.\scripts\run_migrations.ps1
```

## Migration Files

All migration files are located in:
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

## Database Schema Overview

### Core Tables

**super_admins** - Platform administrators
- id (UUID)
- email
- password_hash
- role
- is_active

**organisations** - Tenant organisations
- id (UUID)
- name
- slug
- email
- status

**users** - Organisation users
- id (UUID)
- organisation_id (UUID)
- email
- name
- role (admin/staff)

**customers** - Organisation customers
- id (UUID)
- organisation_id (UUID)
- name
- email
- phone

**invoices** - Customer invoices
- id (UUID)
- organisation_id (UUID)
- customer_id (UUID)
- invoice_number
- status
- total

**payments** - Invoice payments
- id (UUID)
- invoice_id (UUID)
- amount
- method

## Troubleshooting

### Error: "goose: command not found"

Install goose:
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Make sure `$GOPATH/bin` is in your PATH:
```bash
# Windows PowerShell
$env:PATH += ";$env:USERPROFILE\go\bin"

# Linux/Mac
export PATH=$PATH:$HOME/go/bin
```

### Error: "connection refused"

1. Check your DATABASE_URL is correct
2. Verify your Neon database is active
3. Check network connectivity
4. Ensure SSL mode is set correctly

### Error: "permission denied"

For Neon databases, ensure:
1. You're using the correct user (usually `neondb_owner`)
2. The password is correct
3. SSL mode is set to `require`

### Error: "database does not exist"

Neon creates the database automatically. If you see this error:
1. Check the database name in your connection string
2. Verify you're connecting to the right project
3. Try creating a new database in Neon console

### Error: "migration already applied"

This is normal if you've run migrations before. Check status:
```bash
goose -dir internal/db/migrations postgres "$DATABASE_URL" status
```

### Error: "UUID extension not found"

The first migration creates the UUID extension. If it fails:
```sql
-- Run this manually in your database
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

## Verify Database Setup

After running migrations, verify tables exist:

### Using psql
```bash
psql "$DATABASE_URL" -c "\dt"
```

### Using Neon Console
1. Go to https://console.neon.tech
2. Select your project
3. Click "Tables" in the sidebar
4. You should see all 19 tables

### Using the API
```bash
# Start the server
go run cmd/server/main.go

# Test database connection
curl http://localhost:8080/ready
```

## Next Steps

After successful migration:

1. ✅ Create SuperAdmin
   ```powershell
   .\scripts\create_superadmin.ps1
   ```

2. ✅ Start the server
   ```bash
   go run cmd/server/main.go
   ```

3. ✅ Test the API
   ```bash
   curl http://localhost:8080/health
   ```

## Migration Best Practices

1. **Always backup** before running migrations in production
2. **Test migrations** in development first
3. **Review migration files** before running
4. **Check status** after running migrations
5. **Never edit** applied migration files
6. **Create new migrations** for schema changes

## Creating New Migrations

```bash
# Create a new migration file
goose -dir internal/db/migrations create add_new_table sql

# This creates a new file like:
# 020_add_new_table.sql
```

Edit the file:
```sql
-- +goose Up
CREATE TABLE new_table (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS new_table;
```

Then run:
```bash
goose -dir internal/db/migrations postgres "$DATABASE_URL" up
```

## Database Maintenance

### Check Connection Pool
```bash
# In psql
SELECT * FROM pg_stat_activity WHERE datname = 'your_database';
```

### View Table Sizes
```bash
# In psql
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

---

**Need Help?**

- Check `COMPLETE_SETUP_GUIDE.md` for full setup
- See `START_HERE.md` for quick start
- Review migration files in `internal/db/migrations/`
