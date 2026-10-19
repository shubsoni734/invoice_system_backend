package common

import "testing"

func TestTokenMatches(t *testing.T) {
	if tokenMatches("", "") || tokenMatches("secret", "") || tokenMatches("", "secret") {
		t.Fatal("empty values must not match")
	}
	if tokenMatches("secret", "wrong") {
		t.Fatal("different tokens must not match")
	}
	if !tokenMatches("secret", "secret") || !tokenMatches(" secret ", "secret") {
		t.Fatal("matching tokens must be accepted")
	}
}
