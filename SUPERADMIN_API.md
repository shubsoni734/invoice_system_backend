# SuperAdmin API Documentation

## Create SuperAdmin (POST)

### Endpoint
```http
POST /superadmin/auth/create
Content-Type: application/json
```

### Description
Creates a new SuperAdmin account. This endpoint does NOT require authentication, allowing you to create the first superadmin.

### Request Body
```json
{
  "email": "superadmin@invoicepro.com",
  "password": "SuperAdmin@123",
  "role": "superadmin"
}
```

### Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| email | string | Yes | Valid email address |
| password | string | Yes | Minimum 8 characters |
| role | string | No | Default: "superadmin" |

### Success Response (201 Created)
```json
{
  "success": true,
  "message": "SuperAdmin created successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "superadmin@invoicepro.com",
    "role": "superadmin"
  }
}
```

### Error Responses

#### 400 Bad Request - Invalid Input
```json
{
  "success": false,
  "message": "Invalid request: email is required"
}
```

#### 409 Conflict - Email Already Exists
```json
{
  "success": false,
  "message": "SuperAdmin with this email already exists"
}
```

#### 500 Internal Server Error
```json
{
  "success": false,
  "message": "Failed to create superadmin"
}
```

---

## SuperAdmin Login

### Endpoint
```http
POST /superadmin/auth/login
Content-Type: application/json
```

### Request Body
```json
{
  "email": "superadmin@invoicepro.com",
  "password": "SuperAdmin@123"
}
```

### Success Response (200 OK)
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "superadmin@invoicepro.com",
      "role": "superadmin"
    }
  }
}
```

### Error Responses

#### 401 Unauthorized - Invalid Credentials
```json
{
  "success": false,
  "message": "Invalid email or password"
}
```

#### 403 Forbidden - Account Locked
```json
{
  "success": false,
  "message": "Account is locked. Try again later"
}
```

#### 403 Forbidden - Account Inactive
```json
{
  "success": false,
  "message": "Account is inactive"
}
```

---

## Get Current SuperAdmin

### Endpoint
```http
GET /superadmin/auth/me
Authorization: Bearer <access_token>
```

### Success Response (200 OK)
```json
{
  "success": true,
  "message": "SuperAdmin retrieved",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "superadmin@invoicepro.com",
    "role": "superadmin",
    "is_active": true,
    "created_at": "2024-03-10T10:30:00Z"
  }
}
```

---

## Logout

### Endpoint
```http
POST /superadmin/auth/logout
Authorization: Bearer <access_token>
```

### Success Response (200 OK)
```json
{
  "success": true,
  "message": "Logout successful",
  "data": null
}
```

---

## Usage Examples

### Using curl

#### Create SuperAdmin
```bash
curl -X POST http://localhost:8080/superadmin/auth/create \
  -H "Content-Type: application/json" \
  -d '{
    "email": "superadmin@invoicepro.com",
    "password": "SuperAdmin@123"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8080/superadmin/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "superadmin@invoicepro.com",
    "password": "SuperAdmin@123"
  }'
```

#### Get Current User
```bash
curl -X GET http://localhost:8080/superadmin/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

#### Logout
```bash
curl -X POST http://localhost:8080/superadmin/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Using PowerShell

#### Create SuperAdmin
```powershell
$body = @{
    email = "superadmin@invoicepro.com"
    password = "SuperAdmin@123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/superadmin/auth/create" `
    -Method Post `
    -ContentType "application/json" `
    -Body $body
```

#### Login
```powershell
$body = @{
    email = "superadmin@invoicepro.com"
    password = "SuperAdmin@123"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8080/superadmin/auth/login" `
    -Method Post `
    -ContentType "application/json" `
    -Body $body

# Save token
$token = $response.data.access_token
```

#### Get Current User
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/superadmin/auth/me" `
    -Method Get `
    -Headers @{ Authorization = "Bearer $token" }
```

---

## Security Features

### Password Requirements
- Minimum 8 characters
- Hashed using bcrypt with default cost (10)

### Account Lockout
- Account locks after 5 failed login attempts
- Lockout duration: 15 minutes
- Failed attempts reset on successful login

### Token Security
- Access tokens expire after 15 minutes
- Refresh tokens expire after 24 hours
- Tokens use RS256 (RSA) encryption
- Refresh tokens stored with hash in database

### IP Allowlist
- SuperAdmin access can be restricted by IP
- Configure in `.env`: `SA_IP_ALLOWLIST=127.0.0.1,your.ip.address`

---

## Rate Limiting

- Auth endpoints: 10 requests per minute
- Prevents brute force attacks
- Returns 429 Too Many Requests when exceeded

---

## Complete Workflow

### 1. Create First SuperAdmin
```bash
curl -X POST http://localhost:8080/superadmin/auth/create \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "SecurePassword123!"
  }'
```

### 2. Login
```bash
curl -X POST http://localhost:8080/superadmin/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "SecurePassword123!"
  }'
```

Save the `access_token` from the response.

### 3. Use Token for Protected Endpoints
```bash
curl -X GET http://localhost:8080/superadmin/organisations \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

---

## Testing with Postman

### 1. Create SuperAdmin Request

**Method:** POST  
**URL:** `{{base_url}}/superadmin/auth/create`  
**Headers:**
```
Content-Type: application/json
```
**Body (raw JSON):**
```json
{
  "email": "superadmin@invoicepro.com",
  "password": "SuperAdmin@123"
}
```

### 2. Login Request

**Method:** POST  
**URL:** `{{base_url}}/superadmin/auth/login`  
**Headers:**
```
Content-Type: application/json
```
**Body (raw JSON):**
```json
{
  "email": "superadmin@invoicepro.com",
  "password": "SuperAdmin@123"
}
```

**Tests (to save token):**
```javascript
if (pm.response.code === 200) {
    var jsonData = pm.response.json();
    pm.collectionVariables.set('super_access_token', jsonData.data.access_token);
    pm.collectionVariables.set('refresh_token', jsonData.data.refresh_token);
}
```

### 3. Get Current User Request

**Method:** GET  
**URL:** `{{base_url}}/superadmin/auth/me`  
**Headers:**
```
Authorization: Bearer {{super_access_token}}
```

---

## Best Practices

1. **Change Default Credentials**: Never use default passwords in production
2. **Use Strong Passwords**: Minimum 12 characters with mixed case, numbers, and symbols
3. **Secure Storage**: Store credentials in a password manager
4. **IP Allowlist**: Restrict SuperAdmin access to known IPs in production
5. **HTTPS Only**: Always use HTTPS in production
6. **Monitor Access**: Review audit logs regularly
7. **Rotate Tokens**: Implement token rotation for long-lived sessions

---

## Troubleshooting

### Issue: "SuperAdmin with this email already exists"
**Solution:** Use a different email or login with existing credentials

### Issue: "Account is locked"
**Solution:** Wait 15 minutes or contact system administrator

### Issue: "Invalid email or password"
**Solution:** Check credentials are correct, case-sensitive

### Issue: "Failed to create superadmin"
**Solution:** Check database connection and migrations are run

---

## Next Steps

After creating SuperAdmin:

1. ✅ Login to get access token
2. ✅ Create organisations
3. ✅ Create plans
4. ✅ Create organisation admin users
5. ✅ Start using the system

---

**Need Help?**

- Check `API_ENDPOINTS.md` for all API endpoints
- See `COMPLETE_SETUP_GUIDE.md` for setup instructions
- Review `START_HERE.md` for quick start
