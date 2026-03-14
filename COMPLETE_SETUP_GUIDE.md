# Complete Setup Guide - InvoicePro Backend

## 📋 Prerequisites

- Go 1.22 or higher
- PostgreSQL 14+ (or Neon Database)
- Git

## 🚀 Quick Start (5 Minutes)

### Step 1: Clone and Setup

```bash
cd "BillingSystem Backend"
```

### Step 2: Install Dependencies

```bash
go mod download
```

### Step 3: Configure Environment

Copy `.env.example` to `.env` and update:

```bash
cp .env.example .env
```

Edit `.env`:
```env
DATABASE_URL=your_postgresql_connection_string
SERVER_PORT=8080
ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173,http://localhost:8081
```

### Step 4: Generate JWT Keys

```bash
go run generate_keys.go
```

This creates RSA key pairs in the `keys/` directory.

### Step 5: Run Database Migrations

Install goose (migration tool):
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Run migrations:
```bash
goose -dir internal/db/migrations postgres "YOUR_DATABASE_URL" up
```

Or use the setup script:
```powershell
# Windows
.\setup.ps1

# Linux/Mac
chmod +x setup.sh
./setup.sh
```

### Step 6: Create SuperAdmin

```powershell
# Windows
.\scripts\create_superadmin.ps1

# Linux/Mac
chmod +x ./scripts/create_superadmin.sh
./scripts/create_superadmin.sh
```

Default credentials:
- Email: `superadmin@invoicepro.com`
- Password: `SuperAdmin@123`

### Step 7: Start the Server

```bash
go run cmd/server/main.go
```

Server will start on `http://localhost:8080`

### Step 8: Test the API

```bash
# Health check
curl http://localhost:8080/health

# SuperAdmin login
curl -X POST http://localhost:8080/superadmin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"superadmin@invoicepro.com","password":"SuperAdmin@123"}'
```

## 📚 Detailed Setup

### Database Setup

#### Option 1: Local PostgreSQL

1. Install PostgreSQL
2. Create database:
```sql
CREATE DATABASE invoicepro;
```

3. Update `.env`:
```env
DATABASE_URL=postgresql://postgres:password@localhost:5432/invoicepro?sslmode=disable
```

#### Option 2: Neon Database (Cloud)

1. Sign up at https://neon.tech
2. Create a new project
3. Copy connection string to `.env`

### Migration Commands

```bash
# Run all migrations
goose -dir internal/db/migrations postgres "$DATABASE_URL" up

# Rollback last migration
goose -dir internal/db/migrations postgres "$DATABASE_URL" down

# Check migration status
goose -dir internal/db/migrations postgres "$DATABASE_URL" status

# Create new migration
goose -dir internal/db/migrations create migration_name sql
```

### Environment Variables Explained

```env
# Server Configuration
SERVER_PORT=8080                    # Port to run server on
ENVIRONMENT=development             # development or production
ALLOWED_ORIGINS=http://localhost:*  # CORS allowed origins

# Database
DATABASE_URL=postgresql://...       # PostgreSQL connection string
DB_MIN_CONNS=5                      # Minimum connections in pool
DB_MAX_CONNS=25                     # Maximum connections in pool

# JWT for Organisation Users
ORG_JWT_PRIVATE_KEY_PATH=./keys/org_private.pem
ORG_JWT_PUBLIC_KEY_PATH=./keys/org_public.pem
ORG_ACCESS_TOKEN_EXPIRY=15m         # Access token lifetime
ORG_REFRESH_TOKEN_EXPIRY=168h      # Refresh token lifetime (7 days)

# JWT for SuperAdmin
SA_JWT_PRIVATE_KEY_PATH=./keys/sa_private.pem
SA_JWT_PUBLIC_KEY_PATH=./keys/sa_public.pem
SA_ACCESS_TOKEN_EXPIRY=15m
SA_REFRESH_TOKEN_EXPIRY=24h

# Security
SA_IP_ALLOWLIST=127.0.0.1          # Comma-separated IPs for SuperAdmin

# Rate Limiting
RATE_LIMIT_AUTH_RPM=10             # Auth requests per minute
RATE_LIMIT_API_RPM=300             # API requests per minute

# File Uploads
UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE_MB=2

# Logging
LOG_LEVEL=info                      # debug, info, warn, error
LOG_FORMAT=json                     # json or console
```

## 🧪 Testing the API

### Using Postman

1. Import collection: `postman/InvoicePro_API_Collection.json`
2. Set base_url variable: `http://localhost:8080`
3. Start with SuperAdmin login
4. Copy access_token to collection variables

### Using curl

See `API_ENDPOINTS.md` for all endpoints.

Example workflow:

```bash
# 1. SuperAdmin Login
curl -X POST http://localhost:8080/superadmin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"superadmin@invoicepro.com","password":"SuperAdmin@123"}'

# Save the access_token from response

# 2. Create Organisation
curl -X POST http://localhost:8080/superadmin/organisations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "Acme Corp",
    "slug": "acme-corp",
    "email": "admin@acme.com"
  }'

# 3. Create Organisation Admin
curl -X POST http://localhost:8080/superadmin/organisations/ORG_ID/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "email": "admin@acme.com",
    "password": "SecurePass123!",
    "name": "Admin User",
    "role": "admin"
  }'

# 4. Organisation User Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@acme.com","password":"SecurePass123!"}'
```

## 🏗️ Project Structure

```
BillingSystem Backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── constants/
│   │   ├── errors.go            # Error constants
│   │   ├── roles.go             # Role constants
│   │   └── status.go            # Status constants
│   ├── db/
│   │   └── migrations/          # Database migrations
│   ├── middleware/
│   │   ├── auth.go              # Authentication middleware
│   │   ├── cors.go              # CORS middleware
│   │   ├── logger.go            # Logging middleware
│   │   ├── ratelimit.go         # Rate limiting
│   │   └── ...
│   └── utils/
│       ├── hash.go              # Password hashing
│       ├── jwt.go               # JWT utilities
│       ├── response.go          # Response helpers
│       └── ...
├── keys/                        # JWT key pairs
├── scripts/
│   ├── create_superadmin.go     # SuperAdmin creation script
│   ├── create_superadmin.ps1    # PowerShell wrapper
│   └── create_superadmin.sh     # Bash wrapper
├── uploads/                     # File uploads directory
├── .env                         # Environment variables
├── go.mod                       # Go dependencies
└── README.md                    # Documentation
```

## 🔒 Security Best Practices

1. **Change Default Credentials**: Never use default passwords in production
2. **Use HTTPS**: Always use HTTPS in production
3. **Secure Keys**: Keep JWT keys secure and never commit to git
4. **IP Allowlist**: Configure SuperAdmin IP allowlist
5. **Rate Limiting**: Adjust rate limits based on your needs
6. **Database Security**: Use strong database passwords
7. **Environment Variables**: Never commit `.env` file

## 🐛 Troubleshooting

### Server won't start

```bash
# Check if port is already in use
netstat -ano | findstr :8080

# Kill process using the port (Windows)
taskkill /PID <PID> /F

# Linux/Mac
lsof -ti:8080 | xargs kill -9
```

### Database connection fails

1. Check DATABASE_URL is correct
2. Verify database is running
3. Check network connectivity
4. Verify SSL mode (use `sslmode=disable` for local)

### Migrations fail

```bash
# Check migration status
goose -dir internal/db/migrations postgres "$DATABASE_URL" status

# Reset database (⚠️ DESTRUCTIVE)
goose -dir internal/db/migrations postgres "$DATABASE_URL" reset
goose -dir internal/db/migrations postgres "$DATABASE_URL" up
```

### JWT errors

1. Ensure keys are generated: `go run generate_keys.go`
2. Check key paths in `.env`
3. Verify file permissions on key files

## 📖 Additional Resources

- **API Documentation**: `API_ENDPOINTS.md`
- **SuperAdmin Setup**: `SUPERADMIN_SETUP.md`
- **Quick Start**: `QUICKSTART.md`
- **Technology Stack**: `technology-stack.md`

## 🚀 Production Deployment

### Environment Setup

```env
ENVIRONMENT=production
SERVER_PORT=8080
ALLOWED_ORIGINS=https://yourdomain.com

# Use production database
DATABASE_URL=postgresql://...

# Increase connection pool
DB_MAX_CONNS=50

# Secure SuperAdmin access
SA_IP_ALLOWLIST=your.office.ip,another.ip

# Production logging
LOG_LEVEL=warn
LOG_FORMAT=json
```

### Build for Production

```bash
# Build binary
go build -o invoicepro-server cmd/server/main.go

# Run binary
./invoicepro-server
```

### Using Docker

```bash
# Build image
docker build -t invoicepro-backend .

# Run container
docker run -p 8080:8080 --env-file .env invoicepro-backend
```

## 🎯 Next Steps

1. ✅ Complete setup (you're here!)
2. ✅ Create SuperAdmin
3. ✅ Test API endpoints
4. ✅ Create organisations and users
5. ✅ Connect frontend application
6. ✅ Implement remaining API endpoints
7. ✅ Add business logic
8. ✅ Write tests
9. ✅ Deploy to production

---

**Need Help?**

- Check `API_ENDPOINTS.md` for API documentation
- See `SUPERADMIN_SETUP.md` for SuperAdmin details
- Review `technology-stack.md` for architecture info
