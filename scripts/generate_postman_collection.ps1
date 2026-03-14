# Generate Postman Collection
Write-Host "Generating Postman Collection..." -ForegroundColor Cyan

$collection = @"
{
  "info": {
    "name": "InvoicePro API",
    "description": "Complete API collection for InvoicePro Billing System",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {"key": "base_url", "value": "http://localhost:8080", "type": "string"},
    {"key": "access_token", "value": "", "type": "string"},
    {"key": "refresh_token", "value": "", "type": "string"},
    {"key": "super_access_token", "value": "", "type": "string"},
    {"key": "organisation_id", "value": "", "type": "string"}
  ],
  "item": [
    {
      "name": "Health & Status",
      "item": [
        {
          "name": "Health Check",
          "request": {
            "method": "GET",
            "header": [],
            "url": {"raw": "{{base_url}}/health", "host": ["{{base_url}}"], "path": ["health"]}
          }
        },
        {
          "name": "Ready Check",
          "request": {
            "method": "GET",
            "header": [],
            "url": {"raw": "{{base_url}}/ready", "host": ["{{base_url}}"], "path": ["ready"]}
          }
        }
      ]
    },
    {
      "name": "SuperAdmin Auth",
      "item": [
        {
          "name": "SuperAdmin Login",
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "if (pm.response.code === 200) {",
                  "    var jsonData = pm.response.json();",
                  "    pm.collectionVariables.set('super_access_token', jsonData.data.access_token);",
                  "    pm.collectionVariables.set('refresh_token', jsonData.data.refresh_token);",
                  "}"
                ]
              }
            }
          ],
          "request": {
            "method": "POST",
            "header": [{"key": "Content-Type", "value": "application/json"}],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"superadmin@invoicepro.com\",\n  \"password\": \"SuperAdmin@123\"\n}"
            },
            "url": {"raw": "{{base_url}}/superadmin/auth/login", "host": ["{{base_url}}"], "path": ["superadmin", "auth", "login"]}
          }
        }
      ]
    },
    {
      "name": "Organisation Auth",
      "item": [
        {
          "name": "Organisation Login",
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "if (pm.response.code === 200) {",
                  "    var jsonData = pm.response.json();",
                  "    pm.collectionVariables.set('access_token', jsonData.data.access_token);",
                  "    pm.collectionVariables.set('refresh_token', jsonData.data.refresh_token);",
                  "    pm.collectionVariables.set('organisation_id', jsonData.data.user.organisation_id);",
                  "}"
                ]
              }
            }
          ],
          "request": {
            "method": "POST",
            "header": [{"key": "Content-Type", "value": "application/json"}],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"admin@example.com\",\n  \"password\": \"password123\"\n}"
            },
            "url": {"raw": "{{base_url}}/api/v1/auth/login", "host": ["{{base_url}}"], "path": ["api", "v1", "auth", "login"]}
          }
        },
        {
          "name": "Get Current User",
          "request": {
            "method": "GET",
            "header": [{"key": "Authorization", "value": "Bearer {{access_token}}"}],
            "url": {"raw": "{{base_url}}/api/v1/auth/me", "host": ["{{base_url}}"], "path": ["api", "v1", "auth", "me"]}
          }
        }
      ]
    },
    {
      "name": "Customers",
      "item": [
        {
          "name": "Get All Customers",
          "request": {
            "method": "GET",
            "header": [{"key": "Authorization", "value": "Bearer {{access_token}}"}],
            "url": {"raw": "{{base_url}}/api/v1/customers?page=1&limit=10", "host": ["{{base_url}}"], "path": ["api", "v1", "customers"], "query": [{"key": "page", "value": "1"}, {"key": "limit", "value": "10"}]}
          }
        },
        {
          "name": "Create Customer",
          "request": {
            "method": "POST",
            "header": [{"key": "Authorization", "value": "Bearer {{access_token}}"}, {"key": "Content-Type", "value": "application/json"}],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"name\": \"ABC Company\",\n  \"email\": \"contact@abc.com\",\n  \"phone\": \"+1234567890\",\n  \"address\": \"123 Business St\"\n}"
            },
            "url": {"raw": "{{base_url}}/api/v1/customers", "host": ["{{base_url}}"], "path": ["api", "v1", "customers"]}
          }
        }
      ]
    },
    {
      "name": "Invoices",
      "item": [
        {
          "name": "Get All Invoices",
          "request": {
            "method": "GET",
            "header": [{"key": "Authorization", "value": "Bearer {{access_token}}"}],
            "url": {"raw": "{{base_url}}/api/v1/invoices?page=1&limit=10", "host": ["{{base_url}}"], "path": ["api", "v1", "invoices"], "query": [{"key": "page", "value": "1"}, {"key": "limit", "value": "10"}]}
          }
        },
        {
          "name": "Create Invoice",
          "request": {
            "method": "POST",
            "header": [{"key": "Authorization", "value": "Bearer {{access_token}}"}, {"key": "Content-Type", "value": "application/json"}],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"customer_id\": \"uuid\",\n  \"issued_date\": \"2024-03-10\",\n  \"due_date\": \"2024-04-10\",\n  \"items\": [\n    {\n      \"description\": \"Service\",\n      \"quantity\": 1,\n      \"unit_price\": 100,\n      \"tax_rate\": 18\n    }\n  ]\n}"
            },
            "url": {"raw": "{{base_url}}/api/v1/invoices", "host": ["{{base_url}}"], "path": ["api", "v1", "invoices"]}
          }
        }
      ]
    }
  ]
}
"@

$outputPath = "postman\InvoicePro_API_Collection.json"
$collection | Out-File -FilePath $outputPath -Encoding UTF8

Write-Host "✅ Postman collection generated: $outputPath" -ForegroundColor Green
Write-Host ""
Write-Host "Import this file into Postman to start testing the API!" -ForegroundColor Cyan
