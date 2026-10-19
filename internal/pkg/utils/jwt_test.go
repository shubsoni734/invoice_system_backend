package utils

import (
	"testing"
	"time"
)

func TestJWTGenerateAndVerify(t *testing.T) {
	manager, err := NewJWTManager("test-secret-that-is-long-enough", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := manager.GenerateToken(&Claims{UserID: "user-1", OrgID: "org-1", Role: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := manager.VerifyToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "user-1" || claims.OrgID != "org-1" || claims.Role != "admin" {
		t.Fatalf("claims = %+v", claims)
	}
}

func TestJWTRejectsEmptySecret(t *testing.T) {
	if _, err := NewJWTManager("", time.Hour); err == nil {
		t.Fatal("expected empty secret to fail")
	}
}

func TestJWTRejectsWrongSecretAndExpiredToken(t *testing.T) {
	manager, err := NewJWTManager("correct-secret-value-32-characters", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := manager.GenerateToken(&Claims{UserID: "user-1"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := NewJWTManager("different-secret-value-32-chars!", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.VerifyToken(token); err == nil {
		t.Fatal("expected wrong secret to fail")
	}

	expired, err := NewJWTManager("correct-secret-value-32-characters", -time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	old, err := expired.GenerateToken(&Claims{UserID: "user-1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := expired.VerifyToken(old); err == nil {
		t.Fatal("expected expired token to fail")
	}
}

func TestJWTRejectsUnexpectedSigningMethod(t *testing.T) {
	manager, err := NewJWTManager("correct-secret-value-32-characters", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	noneToken := "eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoidXNlci0xIn0."
	if _, err := manager.VerifyToken(noneToken); err == nil {
		t.Fatal("expected unsigned token to fail")
	}
}
