#!/bin/bash

# Bash script to create superadmin
# Usage: ./scripts/create_superadmin.sh

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  SuperAdmin Creation Script"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "❌ Error: .env file not found!"
    echo "Please create a .env file with DATABASE_URL"
    exit 1
fi

# Run the Go script
echo "Creating SuperAdmin..."
echo ""

go run ./scripts/create_superadmin.go

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Script completed successfully!"
else
    echo ""
    echo "❌ Script failed with error code: $?"
fi
