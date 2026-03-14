#!/bin/bash

# Run Database Migrations
# Usage: ./run_migrations.sh [up|down|status|reset]

ACTION=${1:-up}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Database Migration Script"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "⚠️  Warning: .env file not found!"
    echo "Using .env.example instead..."
    
    if [ -f ".env.example" ]; then
        cp .env.example .env
        echo "✅ Created .env from .env.example"
    else
        echo "❌ Error: Neither .env nor .env.example found!"
        exit 1
    fi
fi

# Load environment variables from .env
echo "Loading environment variables..."
export $(grep -v '^#' .env | xargs)

if [ -z "$DATABASE_URL" ]; then
    echo "❌ Error: DATABASE_URL not found in .env file!"
    echo "Please add DATABASE_URL to your .env file"
    exit 1
fi

echo "✅ Database URL loaded"
echo ""

# Check if goose is installed
echo "Checking for goose..."
if ! command -v goose &> /dev/null; then
    echo "❌ goose is not installed!"
    echo ""
    echo "Installing goose..."
    go install github.com/pressly/goose/v3/cmd/goose@latest
    
    if [ $? -eq 0 ]; then
        echo "✅ goose installed successfully!"
    else
        echo "❌ Failed to install goose"
        exit 1
    fi
else
    echo "✅ goose is installed"
fi

echo ""

# Run migration based on action
case "$ACTION" in
    up)
        echo "Running migrations UP..."
        echo ""
        goose -dir internal/db/migrations postgres "$DATABASE_URL" up
        ;;
    down)
        echo "⚠️  WARNING: This will rollback the last migration!"
        read -p "Are you sure? (yes/no): " confirm
        
        if [ "$confirm" = "yes" ] || [ "$confirm" = "y" ]; then
            echo ""
            echo "Rolling back last migration..."
            goose -dir internal/db/migrations postgres "$DATABASE_URL" down
        else
            echo "Operation cancelled."
        fi
        ;;
    status)
        echo "Checking migration status..."
        echo ""
        goose -dir internal/db/migrations postgres "$DATABASE_URL" status
        ;;
    reset)
        echo "⚠️  WARNING: This will DROP ALL TABLES and re-run migrations!"
        echo "This is DESTRUCTIVE and cannot be undone!"
        read -p "Are you ABSOLUTELY sure? (type 'yes' to confirm): " confirm
        
        if [ "$confirm" = "yes" ]; then
            echo ""
            echo "Resetting database..."
            goose -dir internal/db/migrations postgres "$DATABASE_URL" reset
            
            echo ""
            echo "Re-running migrations..."
            goose -dir internal/db/migrations postgres "$DATABASE_URL" up
        else
            echo "Operation cancelled."
        fi
        ;;
    redo)
        echo "Redoing last migration..."
        echo ""
        goose -dir internal/db/migrations postgres "$DATABASE_URL" redo
        ;;
    create)
        read -p "Enter migration name: " migration_name
        
        if [ -n "$migration_name" ]; then
            echo ""
            echo "Creating migration: $migration_name"
            goose -dir internal/db/migrations create "$migration_name" sql
        else
            echo "❌ Migration name is required!"
        fi
        ;;
    *)
        echo "❌ Unknown action: $ACTION"
        echo ""
        echo "Available actions:"
        echo "  up      - Run all pending migrations"
        echo "  down    - Rollback last migration"
        echo "  status  - Check migration status"
        echo "  reset   - Reset database (DESTRUCTIVE)"
        echo "  redo    - Redo last migration"
        echo "  create  - Create new migration"
        echo ""
        echo "Usage: ./run_migrations.sh [action]"
        exit 1
        ;;
esac

echo ""

if [ $? -eq 0 ]; then
    echo "✅ Migration completed successfully!"
else
    echo "❌ Migration failed with error code: $?"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
