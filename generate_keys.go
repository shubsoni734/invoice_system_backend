package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	// Create keys directory if it doesn't exist
	if err := os.MkdirAll("keys", 0755); err != nil {
		fmt.Printf("Error creating keys directory: %v\n", err)
		return
	}

	// Generate Organisation JWT keys
	fmt.Println("Generating Organisation JWT keys...")
	if err := generateKeyPair("keys/org_private.pem", "keys/org_public.pem"); err != nil {
		fmt.Printf("Error generating org keys: %v\n", err)
		return
	}

	// Generate SuperAdmin JWT keys
	fmt.Println("Generating SuperAdmin JWT keys...")
	if err := generateKeyPair("keys/sa_private.pem", "keys/sa_public.pem"); err != nil {
		fmt.Printf("Error generating superadmin keys: %v\n", err)
		return
	}

	fmt.Println("✅ All JWT keys generated successfully!")
	fmt.Println("Keys created:")
	fmt.Println("  - keys/org_private.pem")
	fmt.Println("  - keys/org_public.pem")
	fmt.Println("  - keys/sa_private.pem")
	fmt.Println("  - keys/sa_public.pem")
}

func generateKeyPair(privateKeyPath, publicKeyPath string) error {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Save private key
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer privateKeyFile.Close()

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}

	// Save public key
	publicKeyFile, err := os.Create(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create public key file: %w", err)
	}
	defer publicKeyFile.Close()

	publicKeyPKIX, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyPKIX,
	}

	if err := pem.Encode(publicKeyFile, publicKeyPEM); err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}

	return nil
}