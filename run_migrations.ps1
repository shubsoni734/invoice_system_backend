# Run Database Migrations
# Usage: .\run_migrations.ps1 [up|down|status|reset]

param(
    [string]$Action = "up"
)

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "  Database Migration Script" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-Host "⚠️  Warning: .env file not found!" -ForegroundColor Yellow
    Write-Host "Using .env.example instead..." -ForegroundColor Yellow
    
    if (Test-Path ".env.example") {
        Copy-Item ".env.example" ".env"
        Write-Host "✅ Created .env from .env.example" -ForegroundColor Green
    } else {
        Write-Host "❌ Error: Neither .env nor .env.example found!" -ForegroundColor Red
        exit 1
    }
}

# Load environment variables from .env
Write-Host "Loading environment variables..." -ForegroundColor Yellow
Get-Content .env | ForEach-Object {
    if ($_ -match '^([^=]+)=(.*)$' -and -not $_.StartsWith('#')) {
        $key = $matches[1].Trim()
        $value = $matches[2].Trim()
        [Environment]::SetEnvironmentVariable($key, $value, "Process")
    }
}

$DATABASE_URL = $env:DATABASE_URL

if (-not $DATABASE_URL) {
    Write-Host "❌ Error: DATABASE_URL not found in .env file!" -ForegroundColor Red
    Write-Host "Please add DATABASE_URL to your .env file" -ForegroundColor Yellow
    exit 1
}

Write-Host "✅ Database URL loaded" -ForegroundColor Green
Write-Host ""

# Check if goose is installed
Write-Host "Checking for goose..." -ForegroundColor Yellow
$gooseInstalled = Get-Command goose -ErrorAction SilentlyContinue

if (-not $gooseInstalled) {
    Write-Host "❌ goose is not installed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Installing goose..." -ForegroundColor Yellow
    go install github.com/pressly/goose/v3/cmd/goose@latest
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ goose installed successfully!" -ForegroundColor Green
    } else {
        Write-Host "❌ Failed to install goose" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "✅ goose is installed" -ForegroundColor Green
}

Write-Host ""

# Run migration based on action
switch ($Action.ToLower()) {
    "up" {
        Write-Host "Running migrations UP..." -ForegroundColor Cyan
        Write-Host ""
        goose -dir internal/db/migrations postgres $DATABASE_URL up
    }
    "down" {
        Write-Host "⚠️  WARNING: This will rollback the last migration!" -ForegroundColor Yellow
        Write-Host "Are you sure? (yes/no): " -NoNewline
        $confirm = Read-Host
        
        if ($confirm -eq "yes" -or $confirm -eq "y") {
            Write-Host ""
            Write-Host "Rolling back last migration..." -ForegroundColor Cyan
            goose -dir internal/db/migrations postgres $DATABASE_URL down
        } else {
            Write-Host "Operation cancelled." -ForegroundColor Yellow
        }
    }
    "status" {
        Write-Host "Checking migration status..." -ForegroundColor Cyan
        Write-Host ""
        goose -dir internal/db/migrations postgres $DATABASE_URL status
    }
    "reset" {
        Write-Host "⚠️  WARNING: This will DROP ALL TABLES and re-run migrations!" -ForegroundColor Red
        Write-Host "This is DESTRUCTIVE and cannot be undone!" -ForegroundColor Red
        Write-Host "Are you ABSOLUTELY sure? (type 'yes' to confirm): " -NoNewline
        $confirm = Read-Host
        
        if ($confirm -eq "yes") {
            Write-Host ""
            Write-Host "Resetting database..." -ForegroundColor Red
            goose -dir internal/db/migrations postgres $DATABASE_URL reset
            
            Write-Host ""
            Write-Host "Re-running migrations..." -ForegroundColor Cyan
            goose -dir internal/db/migrations postgres $DATABASE_URL up
        } else {
            Write-Host "Operation cancelled." -ForegroundColor Yellow
        }
    }
    "redo" {
        Write-Host "Redoing last migration..." -ForegroundColor Cyan
        Write-Host ""
        goose -dir internal/db/migrations postgres $DATABASE_URL redo
    }
    "create" {
        Write-Host "Enter migration name: " -NoNewline
        $migrationName = Read-Host
        
        if ($migrationName) {
            Write-Host ""
            Write-Host "Creating migration: $migrationName" -ForegroundColor Cyan
            goose -dir internal/db/migrations create $migrationName sql
        } else {
            Write-Host "❌ Migration name is required!" -ForegroundColor Red
        }
    }
    default {
        Write-Host "❌ Unknown action: $Action" -ForegroundColor Red
        Write-Host ""
        Write-Host "Available actions:" -ForegroundColor Yellow
        Write-Host "  up      - Run all pending migrations"
        Write-Host "  down    - Rollback last migration"
        Write-Host "  status  - Check migration status"
        Write-Host "  reset   - Reset database (DESTRUCTIVE)"
        Write-Host "  redo    - Redo last migration"
        Write-Host "  create  - Create new migration"
        Write-Host ""
        Write-Host "Usage: .\run_migrations.ps1 [action]" -ForegroundColor Cyan
        exit 1
    }
}

Write-Host ""

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Migration completed successfully!" -ForegroundColor Green
} else {
    Write-Host "❌ Migration failed with error code: $LASTEXITCODE" -ForegroundColor Red
}

Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
