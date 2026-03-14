# Database Migration Script for InvoicePro
# This script runs all database migrations

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "  Database Migration Script" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-Host "❌ Error: .env file not found!" -ForegroundColor Red
    Write-Host "Please create a .env file with DATABASE_URL" -ForegroundColor Yellow
    Write-Host "Example: DATABASE_URL=postgresql://user:pass@host:port/db?sslmode=require" -ForegroundColor Yellow
    exit 1
}

# Load environment variables
Get-Content .env | ForEach-Object {
    if ($_ -match '^([^=]+)=(.*)$') {
        $name = $matches[1].Trim()
        $value = $matches[2].Trim()
        [Environment]::SetEnvironmentVariable($name, $value, "Process")
    }
}

$DATABASE_URL = $env:DATABASE_URL

if (-not $DATABASE_URL) {
    Write-Host "❌ Error: DATABASE_URL not found in .env file!" -ForegroundColor Red
    exit 1
}

Write-Host "📊 Database URL: $($DATABASE_URL.Substring(0, [Math]::Min(50, $DATABASE_URL.Length)))..." -ForegroundColor Gray
Write-Host ""

# Check if goose is installed
$gooseInstalled = Get-Command goose -ErrorAction SilentlyContinue

if (-not $gooseInstalled) {
    Write-Host "⚠️  Goose migration tool not found!" -ForegroundColor Yellow
    Write-Host "Installing goose..." -ForegroundColor Cyan
    go install github.com/pressly/goose/v3/cmd/goose@latest
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Failed to install goose" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✅ Goose installed successfully!" -ForegroundColor Green
    Write-Host ""
}

# Check migration status
Write-Host "📋 Checking migration status..." -ForegroundColor Cyan
goose -dir internal/db/migrations postgres $DATABASE_URL status

Write-Host ""
Write-Host "🚀 Running migrations..." -ForegroundColor Cyan
Write-Host ""

# Run migrations
goose -dir internal/db/migrations postgres $DATABASE_URL up

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Migrations completed successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📊 Final migration status:" -ForegroundColor Cyan
    goose -dir internal/db/migrations postgres $DATABASE_URL status
    Write-Host ""
    Write-Host "✅ Database is ready!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "1. Create SuperAdmin: .\scripts\create_superadmin.ps1" -ForegroundColor Cyan
    Write-Host "2. Start server: go run cmd/server/main.go" -ForegroundColor Cyan
} else {
    Write-Host ""
    Write-Host "❌ Migration failed!" -ForegroundColor Red
    Write-Host "Please check the error messages above." -ForegroundColor Yellow
    exit 1
}
