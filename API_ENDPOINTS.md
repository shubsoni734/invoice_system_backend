# InvoicePro API Endpoints

## Base URL
```
http://localhost:8080
```

## Authentication

All authenticated endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <access_token>
```

---

## 1. Health & Status

### Health Check
```http
GET /health
```

**Response:**
```json
{
  "success": true,
  "message": "Server is running",
  "data": null
}
```

### Readiness Check
```http
GET /ready
```

**Response:**
```json
{
  "success": true,
  "message": "Server is ready",
  "data": null
}
```

---

## 2. SuperAdmin Authentication

### SuperAdmin Login
```http
POST /superadmin/auth/login
Content-Type: application/json

{
  "email": "superadmin@invoicepro.com",
  "password": "SuperAdmin@123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "user": {
      "id": "uuid",
      "email": "superadmin@invoicepro.com",
      "role": "superadmin"
    }
  }
}
```

### SuperAdmin Refresh Token
```http
POST /superadmin/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGc..."
}
```

---

## 3. SuperAdmin - Organisation Management

### Create Organisation
```http
POST /superadmin/organisations
Authorization: Bearer <super_access_token>
Content-Type: application/json

{
  "name": "Acme Corporation",
  "slug": "acme-corp",
  "email": "admin@acme.com",
  "phone": "+1234567890",
  "address": "123 Business St, City, Country"
}
```

### Get All Organisations
```http
GET /superadmin/organisations?page=1&limit=10&search=acme
Authorization: Bearer <super_access_token>
```

### Get Organisation by ID
```http
GET /superadmin/organisations/:id
Authorization: Bearer <super_access_token>
```

### Update Organisation
```http
PUT /superadmin/organisations/:id
Authorization: Bearer <super_access_token>
Content-Type: application/json

{
  "name": "Updated Name",
  "status": "active"
}
```

### Delete Organisation
```http
DELETE /superadmin/organisations/:id
Authorization: Bearer <super_access_token>
```

---

## 4. SuperAdmin - Plan Management

### Create Plan
```http
POST /superadmin/plans
Authorization: Bearer <super_access_token>
Content-Type: application/json

{
  "name": "Professional",
  "price_monthly": 49.99,
  "price_yearly": 499.99,
  "max_users": 10,
  "max_customers": 1000,
  "max_invoices_per_month": 500,
  "max_storage_mb": 5000,
  "whatsapp_enabled": true,
  "custom_templates": true,
  "api_access": true
}
```

### Get All Plans
```http
GET /superadmin/plans
Authorization: Bearer <super_access_token>
```

### Update Plan
```http
PUT /superadmin/plans/:id
Authorization: Bearer <super_access_token>
```

---

## 5. Organisation User Authentication

### Organisation User Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@acme.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "user": {
      "id": "uuid",
      "organisation_id": "uuid",
      "email": "admin@acme.com",
      "name": "John Doe",
      "role": "admin"
    }
  }
}
```

### Refresh Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGc..."
}
```

### Get Current User
```http
GET /api/v1/auth/me
Authorization: Bearer <access_token>
```

### Logout
```http
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

---

## 6. User Management (Admin Only)

### Create User
```http
POST /api/v1/users
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "email": "user@acme.com",
  "password": "password123",
  "name": "Jane Smith",
  "role": "staff"
}
```

### Get All Users
```http
GET /api/v1/users?page=1&limit=10&search=jane
Authorization: Bearer <access_token>
```

### Get User by ID
```http
GET /api/v1/users/:id
Authorization: Bearer <access_token>
```

### Update User
```http
PUT /api/v1/users/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Jane Doe",
  "role": "admin",
  "is_active": true
}
```

### Delete User
```http
DELETE /api/v1/users/:id
Authorization: Bearer <access_token>
```

---

## 7. Customer Management

### Create Customer
```http
POST /api/v1/customers
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "ABC Company",
  "email": "contact@abc.com",
  "phone": "+1234567890",
  "address": "456 Customer Ave",
  "tax_number": "TAX123456"
}
```

### Get All Customers
```http
GET /api/v1/customers?page=1&limit=10&search=abc
Authorization: Bearer <access_token>
```

### Get Customer by ID
```http
GET /api/v1/customers/:id
Authorization: Bearer <access_token>
```

### Update Customer
```http
PUT /api/v1/customers/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "ABC Corporation",
  "email": "info@abc.com",
  "is_active": true
}
```

### Delete Customer
```http
DELETE /api/v1/customers/:id
Authorization: Bearer <access_token>
```

---

## 8. Service Management

### Create Service
```http
POST /api/v1/services
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Web Development",
  "description": "Custom web development services",
  "unit_price": 100.00,
  "tax_rate": 18.00,
  "unit": "hour"
}
```

### Get All Services
```http
GET /api/v1/services?page=1&limit=10
Authorization: Bearer <access_token>
```

### Update Service
```http
PUT /api/v1/services/:id
Authorization: Bearer <access_token>
```

### Delete Service
```http
DELETE /api/v1/services/:id
Authorization: Bearer <access_token>
```

---

## 9. Invoice Management

### Create Invoice
```http
POST /api/v1/invoices
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "customer_id": "uuid",
  "issued_date": "2024-03-10",
  "due_date": "2024-04-10",
  "items": [
    {
      "service_id": "uuid",
      "description": "Web Development - 10 hours",
      "quantity": 10,
      "unit_price": 100.00,
      "tax_rate": 18.00
    }
  ],
  "discount_amount": 0,
  "notes": "Thank you for your business",
  "terms": "Payment due within 30 days"
}
```

### Get All Invoices
```http
GET /api/v1/invoices?page=1&limit=10&status=draft&from_date=2024-01-01&to_date=2024-12-31
Authorization: Bearer <access_token>
```

### Get Invoice by ID
```http
GET /api/v1/invoices/:id
Authorization: Bearer <access_token>
```

### Update Invoice
```http
PUT /api/v1/invoices/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "status": "sent",
  "notes": "Updated notes"
}
```

### Delete Invoice
```http
DELETE /api/v1/invoices/:id
Authorization: Bearer <access_token>
```

### Send Invoice
```http
POST /api/v1/invoices/:id/send
Authorization: Bearer <access_token>
```

### Mark Invoice as Paid
```http
POST /api/v1/invoices/:id/mark-paid
Authorization: Bearer <access_token>
```

### Download Invoice PDF
```http
GET /api/v1/invoices/:id/pdf
Authorization: Bearer <access_token>
```

---

## 10. Payment Management

### Create Payment
```http
POST /api/v1/payments
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "invoice_id": "uuid",
  "amount": 1180.00,
  "method": "bank_transfer",
  "reference": "TXN123456",
  "notes": "Payment received",
  "payment_date": "2024-03-15"
}
```

### Get All Payments
```http
GET /api/v1/payments?page=1&limit=10&invoice_id=uuid
Authorization: Bearer <access_token>
```

### Get Payment by ID
```http
GET /api/v1/payments/:id
Authorization: Bearer <access_token>
```

---

## 11. Settings Management

### Get Organisation Settings
```http
GET /api/v1/settings
Authorization: Bearer <access_token>
```

### Update Organisation Settings
```http
PUT /api/v1/settings
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "business_name": "Acme Corporation",
  "business_email": "billing@acme.com",
  "business_phone": "+1234567890",
  "business_address": "123 Business St",
  "currency": "USD",
  "default_tax_rate": 18.00,
  "invoice_prefix": "INV",
  "default_due_days": 30
}
```

---

## 12. Dashboard & Reports

### Get Dashboard Stats
```http
GET /api/v1/dashboard/stats
Authorization: Bearer <access_token>
```

### Get Revenue Report
```http
GET /api/v1/reports/revenue?from_date=2024-01-01&to_date=2024-12-31
Authorization: Bearer <access_token>
```

### Get Invoice Report
```http
GET /api/v1/reports/invoices?from_date=2024-01-01&to_date=2024-12-31
Authorization: Bearer <access_token>
```

---

## Error Responses

All endpoints return errors in this format:

```json
{
  "success": false,
  "message": "Error message",
  "errors": {
    "field_name": ["validation error message"]
  }
}
```

### Common HTTP Status Codes

- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Validation error
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `422 Unprocessable Entity` - Validation failed
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

---

## Rate Limits

- Auth endpoints: 10 requests per minute
- API endpoints: 300 requests per minute

---

## Pagination

All list endpoints support pagination:

```
?page=1&limit=10
```

Response includes:
```json
{
  "success": true,
  "data": [...],
  "total": 100,
  "page": 1,
  "limit": 10
}
```
