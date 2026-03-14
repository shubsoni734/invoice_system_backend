package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load environment variables
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	dbURL := viper.GetString("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	fmt.Println("Connected to database successfully!")

	// Get superadmin details from command line or use defaults
	email := getEnvOrPrompt("SUPERADMIN_EMAIL", "superadmin@invoicepro.com")
	password := getEnvOrPrompt("SUPERADMIN_PASSWORD", "SuperAdmin@123")

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v\n", err)
	}

	// Check if superadmin already exists
	var exists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM super_admins WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check if superadmin exists: %v\n", err)
	}

	if exists {
		fmt.Printf("⚠️  SuperAdmin with email '%s' already exists!\n", email)
		fmt.Println("Do you want to update the password? (yes/no)")
		var response string
		fmt.Scanln(&response)
		
		if response == "yes" || response == "y" {
			_, err = pool.Exec(ctx, `
				UPDATE super_admins 
				SET password_hash = $1, updated_at = NOW()
				WHERE email = $2
			`, string(hashedPassword), email)
			
			if err != nil {
				log.Fatalf("Failed to update superadmin: %v\n", err)
			}
			
			fmt.Println("✅ SuperAdmin password updated successfully!")
		} else {
			fmt.Println("Operation cancelled.")
		}
		return
	}

	// Create superadmin
	var id string
	err = pool.QueryRow(ctx, `
		INSERT INTO super_admins (email, password_hash, role, is_active)
		VALUES ($1, $2, 'superadmin', true)
		RETURNING id
	`, email, string(hashedPassword)).Scan(&id)

	if err != nil {
		log.Fatalf("Failed to create superadmin: %v\n", err)
	}

	fmt.Println("\n✅ SuperAdmin created successfully!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📧 Email:    %s\n", email)
	fmt.Printf("🔑 Password: %s\n", password)
	fmt.Printf("🆔 ID:       %s\n", id)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\n⚠️  IMPORTANT: Please save these credentials securely!")
	fmt.Println("You can now use these credentials to login to the SuperAdmin panel.")
}

func getEnvOrPrompt(envKey, defaultValue string) string {
	value := os.Getenv(envKey)
	if value != "" {
		return value
	}
	
	fmt.Printf("Enter %s (default: %s): ", envKey, defaultValue)
	var input string
	fmt.Scanln(&input)
	
	if input == "" {
		return defaultValue
	}
	return input
}
