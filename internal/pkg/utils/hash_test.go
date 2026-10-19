package utils

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("Password123!")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(hash, "Password123!") {
		t.Fatal("expected matching password")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Fatal("expected mismatch")
	}
}

func TestHashTokenIsStableAndNotPlaintext(t *testing.T) {
	first := HashToken("reset-token")
	second := HashToken("reset-token")
	if first != second {
		t.Fatal("token hash should be stable")
	}
	if first == "reset-token" || len(first) != 64 {
		t.Fatalf("hash = %s", first)
	}
	if HashToken("other") == first {
		t.Fatal("different tokens should hash differently")
	}
}

func TestGenerateRandomSecretsHaveRequestedLength(t *testing.T) {
	password := GenerateRandomPassword(16)
	token := GenerateRandomToken(32)
	if len(password) != 16 || len(token) != 32 {
		t.Fatalf("password %d token %d", len(password), len(token))
	}
	if password == GenerateRandomPassword(16) {
		t.Fatal("expected distinct passwords")
	}
}
