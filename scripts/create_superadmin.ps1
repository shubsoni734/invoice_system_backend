# PowerShell script to create superadmin
# Usage: .\scripts\create_superadmin.ps1

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "  SuperAdmin Creation Script" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-Host "❌ Error: .env file not found!" -ForegroundColor Red
    Write-Host "Please create a .env file with DATABASE_URL" -ForegroundColor Yellow
    exit 1
}

# Run the Go script
Write-Host "Creating SuperAdmin..." -ForegroundColor Yellow
Write-Host ""

go run ./scripts/create_superadmin.go

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Script completed successfully!" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "❌ Script failed with error code: $LASTEXITCODE" -ForegroundColor Red
}
