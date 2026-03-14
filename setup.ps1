# Invoice Backend Setup Script
# This script creates all necessary project files

Write-Host "Creating Invoice Backend project structure..." -ForegroundColor Green

# Create .env file
Copy-Item .env.example .env -Force
Write-Host "✓ Created .env file" -ForegroundColor Green

Write-Host "`nSetup complete! Next steps:" -ForegroundColor Yellow
Write-Host "1. Install CLI tools:" -ForegroundColor Cyan
Write-Host "   go install github.com/pressly/goose/v3/cmd/goose@latest"
Write-Host "   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"
Write-Host "`n2. Generate JWT keys:" -ForegroundColor Cyan
Write-Host "   mkdir keys"
Write-Host "   openssl genrsa -out keys/org_private.pem 2048"
Write-Host "   openssl rsa -in keys/org_private.pem -pubout -out keys/org_public.pem"
Write-Host "   openssl genrsa -out keys/sa_private.pem 2048"
Write-Host "   openssl rsa -in keys/sa_private.pem -pubout -out keys/sa_public.pem"
Write-Host "`n3. Run migrations:" -ForegroundColor Cyan
Write-Host "   goose -dir internal/db/migrations postgres `$env:DATABASE_URL up"
Write-Host "`n4. Generate sqlc code:" -ForegroundColor Cyan
Write-Host "   sqlc generate"
Write-Host "`n5. Start server:" -ForegroundColor Cyan
Write-Host "   go run cmd/server/main.go"
