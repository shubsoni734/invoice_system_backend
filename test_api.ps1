# Test API Script
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "  API Testing Script" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080"

# Test 1: Health Check
Write-Host "1. Testing Health Endpoint..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/health" -Method Get
    Write-Host "✅ Health Check: " -ForegroundColor Green -NoNewline
    Write-Host $response.message
} catch {
    Write-Host "❌ Health Check Failed: $_" -ForegroundColor Red
}

Write-Host ""

# Test 2: Create SuperAdmin
Write-Host "2. Creating SuperAdmin..." -ForegroundColor Yellow
$createBody = @{
    email = "superadmin@invoicepro.com"
    password = "SuperAdmin@123"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/superadmin/auth/create" `
        -Method Post `
        -ContentType "application/json" `
        -Body $createBody
    
    Write-Host "✅ SuperAdmin Created!" -ForegroundColor Green
    Write-Host "   ID: $($response.data.id)" -ForegroundColor Cyan
    Write-Host "   Email: $($response.data.email)" -ForegroundColor Cyan
} catch {
    if ($_.Exception.Response.StatusCode -eq 409) {
        Write-Host "⚠️  SuperAdmin already exists (this is OK)" -ForegroundColor Yellow
    } else {
        Write-Host "❌ Create Failed: $_" -ForegroundColor Red
    }
}

Write-Host ""

# Test 3: Login
Write-Host "3. Testing Login..." -ForegroundColor Yellow
$loginBody = @{
    email = "superadmin@invoicepro.com"
    password = "SuperAdmin@123"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/superadmin/auth/login" `
        -Method Post `
        -ContentType "application/json" `
        -Body $loginBody
    
    Write-Host "✅ Login Successful!" -ForegroundColor Green
    $token = $response.data.access_token
    Write-Host "   Token: $($token.Substring(0, 50))..." -ForegroundColor Cyan
    
    # Test 4: Get Current User
    Write-Host ""
    Write-Host "4. Testing Get Current User..." -ForegroundColor Yellow
    
    $headers = @{
        Authorization = "Bearer $token"
    }
    
    $userResponse = Invoke-RestMethod -Uri "$baseUrl/superadmin/auth/me" `
        -Method Get `
        -Headers $headers
    
    Write-Host "✅ User Info Retrieved!" -ForegroundColor Green
    Write-Host "   Email: $($userResponse.data.email)" -ForegroundColor Cyan
    Write-Host "   Role: $($userResponse.data.role)" -ForegroundColor Cyan
    
} catch {
    Write-Host "❌ Login Failed: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "  Testing Complete!" -ForegroundColor Green
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "1. Import Postman collection from postman/ folder"
Write-Host "2. Use the access token to test other endpoints"
Write-Host "3. Check SUPERADMIN_API.md for complete API documentation"
