# 🚀 START HERE - InvoicePro Backend

## Quick Setup (5 Minutes)

### 1. Install Dependencies
```bash
go mod download
```

### 2. Setup Environment
```bash
# Copy example env file
cp .env.example .env

# Edit .env and add your DATABASE_URL
```

### 3. Generate JWT Keys
```bash
go run generate_keys.go
```

### 4. Run Migrations
```bash
# Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run migrations
goose -dir internal/db/migrations postgres "YOUR_DATABASE_URL" up
```

### 5. Create SuperAdmin
```powershell
# Windows
.\scripts\create_superadmin.ps1

# Linux/Mac
chmod +x ./scripts/create_superadmin.sh
./scripts/create_superadmin.sh
```

**Default Credentials:**
- Email: `superadmin@invoicepro.com`
- Password: `SuperAdmin@123`

### 6. Start Server
```bash
go run cmd/server/main.go
```

Server runs on: `http://localhost:8080`

### 7. Test API
```bash
curl http://localhost:8080/health
```

## 📚 Documentation

| Document | Description |
|----------|-------------|
| **COMPLETE_SETUP_GUIDE.md** | Detailed setup instructions |
| **SUPERADMIN_SETUP.md** | SuperAdmin creation guide |
| **API_ENDPOINTS.md** | Complete API documentation |
| **QUICKSTART.md** | Quick start guide |

## 🧪 Testing with Postman

### Generate Postman Collection
```powershell
.\scripts\generate_postman_collection.ps1
```

### Import to Postman
1. Open Postman
2. Click "Import"
3. Select `postman/InvoicePro_API_Collection.json`
4. Start testing!

## 🔑 Default Credentials

### SuperAdmin
```
Email:    superadmin@invoicepro.com
Password: SuperAdmin@123
```

⚠️ **Change these in production!**

## 📋 API Endpoints Overview

### Health & Status
- `GET /health` - Health check
- `GET /ready` - Readiness check

### SuperAdmin
- `POST /superadmin/auth/login` - SuperAdmin login
- `GET /superadmin/organisations` - List organisations
- `POST /superadmin/organisations` - Create organisation

### Organisation Auth
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/auth/me` - Get current user
- `POST /api/v1/auth/refresh` - Refresh token

### Customers
- `GET /api/v1/customers` - List customers
- `POST /api/v1/customers` - Create customer
- `PUT /api/v1/customers/:id` - Update customer
- `DELETE /api/v1/customers/:id` - Delete customer

### Invoices
- `GET /api/v1/invoices` - List invoices
- `POST /api/v1/invoices` - Create invoice
- `POST /api/v1/invoices/:id/send` - Send invoice
- `POST /api/v1/invoices/:id/mark-paid` - Mark as paid
- `GET /api/v1/invoices/:id/pdf` - Download PDF

See `API_ENDPOINTS.md` for complete list.

## 🏗️ Project Structure

```
BillingSystem Backend/
├── cmd/server/main.go           # Entry point
├── internal/
│   ├── config/                  # Configuration
│   ├── db/migrations/           # Database migrations
│   ├── middleware/              # HTTP middleware
│   └── utils/                   # Utilities
├── scripts/
│   ├── create_superadmin.go     # SuperAdmin script
│   └── generate_postman_collection.ps1
├── postman/                     # Postman collections
├── keys/                        # JWT keys
└── .env                         # Environment variables
```

## 🔧 Common Commands

```bash
# Start server
go run cmd/server/main.go

# Run tests
go test ./...

# Build for production
go build -o invoicepro cmd/server/main.go

# Run migrations
goose -dir internal/db/migrations postgres "$DATABASE_URL" up

# Create SuperAdmin
.\scripts\create_superadmin.ps1

# Generate Postman collection
.\scripts\generate_postman_collection.ps1
```

## 🐛 Troubleshooting

### Server won't start
- Check if port 8080 is available
- Verify DATABASE_URL in .env
- Ensure migrations are run

### Database connection fails
- Check DATABASE_URL format
- Verify database is running
- Check network connectivity

### JWT errors
- Run `go run generate_keys.go`
- Check key paths in .env

## 🎯 Next Steps

1. ✅ Complete setup (you're here!)
2. ✅ Test API with Postman
3. ✅ Create organisations
4. ✅ Create organisation users
5. ✅ Connect frontend
6. ✅ Implement business logic
7. ✅ Deploy to production

## 📞 Need Help?

- Check `COMPLETE_SETUP_GUIDE.md` for detailed instructions
- See `API_ENDPOINTS.md` for API documentation
- Review `SUPERADMIN_SETUP.md` for SuperAdmin details

---

**Ready to start?** Follow the Quick Setup above! 🚀
