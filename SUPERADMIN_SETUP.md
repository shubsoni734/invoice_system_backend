# SuperAdmin Setup Guide

## Quick Start

### Prerequisites
1. Go 1.22+ installed
2. PostgreSQL database running
3. `.env` file configured with `DATABASE_URL`

### Create SuperAdmin

#### Option 1: Using PowerShell (Windows)
```powershell
.\scripts\create_superadmin.ps1
```

#### Option 2: Using Bash (Linux/Mac)
```bash
chmod +x ./scripts/create_superadmin.sh
./scripts/create_superadmin.sh
```

#### Option 3: Direct Go Command
```bash
go run ./scripts/create_superadmin.go
```

### Default Credentials

If you don't provide custom credentials, the script will create:

```
Email:    superadmin@invoicepro.com
Password: SuperAdmin@123
```

### Custom Credentials

Set environment variables before running:

```bash
# Windows PowerShell
$env:SUPERADMIN_EMAIL="admin@example.com"
$env:SUPERADMIN_PASSWORD="YourSecurePassword123!"
.\scripts\create_superadmin.ps1

# Linux/Mac
export SUPERADMIN_EMAIL="admin@example.com"
export SUPERADMIN_PASSWORD="YourSecurePassword123!"
./scripts/create_superadmin.sh
```

Or the script will prompt you interactively.

## What the Script Does

1. ✅ Connects to your PostgreSQL database
2. ✅ Checks if superadmin already exists
3. ✅ Hashes the password securely (bcrypt)
4. ✅ Creates superadmin record in `super_admins` table
5. ✅ Displays credentials for you to save

## After Creation

### 1. Test SuperAdmin Login

Use Postman or curl:

```bash
curl -X POST http://localhost:8080/superadmin/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "superadmin@invoicepro.com",
    "password": "SuperAdmin@123"
  }'
```

### 2. Create Your First Organisation

```bash
curl -X POST http://localhost:8080/superadmin/organisations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "name": "My Company",
    "slug": "my-company",
    "email": "admin@mycompany.com",
    "phone": "+1234567890"
  }'
```

### 3. Create Organisation Admin User

```bash
curl -X POST http://localhost:8080/superadmin/organisations/:org_id/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "email": "admin@mycompany.com",
    "password": "SecurePassword123!",
    "name": "Admin User",
    "role": "admin"
  }'
```

## Security Notes

⚠️ **IMPORTANT:**

1. **Change Default Password**: Always change the default password in production
2. **Secure Storage**: Store credentials in a password manager
3. **IP Allowlist**: Configure `SA_IP_ALLOWLIST` in `.env` for production
4. **HTTPS Only**: Use HTTPS in production
5. **Strong Passwords**: Use passwords with:
   - Minimum 12 characters
   - Uppercase and lowercase letters
   - Numbers and special characters

## Troubleshooting

### Error: "Unable to connect to database"

Check your `.env` file:
```env
DATABASE_URL=postgresql://user:password@host:port/database?sslmode=require
```

### Error: "SuperAdmin already exists"

The script will ask if you want to update the password. Type `yes` to proceed.

### Error: "Failed to hash password"

Ensure you have the required Go dependencies:
```bash
go mod download
```

## Database Schema

The `super_admins` table structure:

```sql
CREATE TABLE super_admins (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
```

## Next Steps

1. ✅ Create SuperAdmin (you're here!)
2. ✅ Start the backend server: `go run cmd/server/main.go`
3. ✅ Test API endpoints with Postman
4. ✅ Create organisations and plans
5. ✅ Create organisation admin users
6. ✅ Connect frontend application

## API Documentation

See `API_ENDPOINTS.md` for complete API documentation.

## Postman Collection

Import the Postman collection from `postman/InvoicePro_API_Collection.json` for easy API testing.

---

**Need Help?**

Check the main `README.md` or `QUICKSTART.md` for more information.
