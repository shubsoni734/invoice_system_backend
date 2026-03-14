# Migration Quick Start

## 🚀 Easiest Way (Recommended)

### Windows
```powershell
.\run_migrations.ps1
```

### Linux/Mac
```bash
chmod +x run_migrations.sh
./run_migrations.sh
```

**That's it!** The script will:
- ✅ Check for .env file
- ✅ Load DATABASE_URL
- ✅ Install goose if needed
- ✅ Run all migrations

---

## 📋 Available Commands

### Run Migrations
```powershell
# Windows
.\run_migrations.ps1 up

# Linux/Mac
./run_migrations.sh up
```

### Check Status
```powershell
# Windows
.\run_migrations.ps1 status

# Linux/Mac
./run_migrations.sh status
```

### Rollback Last Migration
```powershell
# Windows
.\run_migrations.ps1 down

# Linux/Mac
./run_migrations.sh down
```

### Create New Migration
```powershell
# Windows
.\run_migrations.ps1 create

# Linux/Mac
./run_migrations.sh create
```

### Reset Database (⚠️ DESTRUCTIVE)
```powershell
# Windows
.\run_migrations.ps1 reset

# Linux/Mac
./run_migrations.sh reset
```

---

## 🔧 Manual Method

If you prefer to run commands manually:

### 1. Install Goose
```powershell
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### 2. Set Database URL
```powershell
# Windows
$env:DATABASE_URL="your_database_url"

# Linux/Mac
export DATABASE_URL="your_database_url"
```

### 3. Run Migrations
```powershell
# Windows
goose -dir internal/db/migrations postgres $env:DATABASE_URL up

# Linux/Mac
goose -dir internal/db/migrations postgres "$DATABASE_URL" up
```

---

## ✅ What Gets Created

Running migrations creates 19 tables:

1. super_admins
2. plans
3. organisations
4. organisation_subscriptions
5. users
6. refresh_tokens
7. super_refresh_tokens
8. impersonation_sessions
9. customers
10. services
11. invoice_sessions
12. invoices
13. invoice_items
14. payments
15. templates
16. whatsapp_logs
17. settings
18. audit_logs
19. super_audit_logs

---

## 🐛 Troubleshooting

### "goose: command not found"
```powershell
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### "DATABASE_URL not found"
Check your `.env` file has:
```env
DATABASE_URL=postgresql://...
```

### "connection refused"
- Check database is running
- Verify DATABASE_URL is correct
- Check network connectivity

---

## 📚 Full Documentation

For detailed information, see:
- `MIGRATION_GUIDE.md` - Complete migration guide
- `COMPLETE_SETUP_GUIDE.md` - Full setup instructions

---

## 🎯 Next Steps

After migrations:

1. ✅ Create SuperAdmin: `.\scripts\create_superadmin.ps1`
2. ✅ Start Server: `go run cmd/server/main.go`
3. ✅ Test API: Import Postman collection

---

**Quick and Easy!** 🚀
