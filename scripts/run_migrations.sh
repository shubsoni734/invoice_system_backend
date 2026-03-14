#!/bin/bash

# Database Migration Script for InvoicePro
# This script runs all database migrations

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Database Migration Script"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "❌ Error: .env file not found!"
    echo "Please create a .env file with DATABASE_URL"
    echo "Example: DATABASE_URL=postgresql://user:pass@host:port/db?sslmode=require"
    exit 1
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)

if [ -z "$DATABASE_URL" ]; then
    echo "❌ Error: DATABASE_URL not found in .env file!"
    exit 1
fi

echo "📊 Database URL: ${DATABASE_URL:0:50}..."
echo ""

# Check if goose is installed
if ! command -v goose &> /dev/null; then
    echo "⚠️  Goose migration tool not found!"
    echo "Installing goose..."
    go install github.com/pressly/goose/v3/cmd/goose@latest
    
    if [ $? -ne 0 ]; then
        echo "❌ Failed to install goose"
        exit 1
    fi
    
    echo "✅ Goose installed successfully!"
    echo ""
fi

# Check migration status
echo "📋 Checking migration status..."
goose -dir internal/db/migrations postgres "$DATABASE_URL" status

echo ""
echo "🚀 Running migrations..."
echo ""

# Run migrations
goose -dir internal/db/migrations postgres "$DATABASE_URL" up

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Migrations completed successfully!"
    echo ""
    echo "📊 Final migration status:"
    goose -dir internal/db/migrations postgres "$DATABASE_URL" status
    echo ""
    echo "✅ Database is ready!"
    echo ""
    echo "Next steps:"
    echo "1. Create SuperAdmin: ./scripts/create_superadmin.sh"
    echo "2. Start server: go run cmd/server/main.go"
else
    echo ""
    echo "❌ Migration failed!"
    echo "Please check the error messages above."
    exit 1
fi
