package config

import (
	"strings"
	"testing"
)

func TestMissingSettings(t *testing.T) {
	missing := missingSettings("", "", "short", "production")
	joined := strings.Join(missing, " ")
	for _, want := range []string{"DATABASE_URL", "ORG_JWT_SECRET", "SA_JWT_SECRET"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %s in %q", want, joined)
		}
	}

	if got := missingSettings("postgres://db", "12345678901234567890123456789012", "12345678901234567890123456789012", "production"); len(got) != 0 {
		t.Fatalf("expected no missing settings, got %v", got)
	}

	short := missingSettings("postgres://db", "too-short", "12345678901234567890123456789012", "production")
	if len(short) != 1 || !strings.Contains(short[0], "ORG_JWT_SECRET") {
		t.Fatalf("short production secret = %v", short)
	}

	dev := missingSettings("postgres://db", "short", "short", "development")
	if len(dev) != 0 {
		t.Fatalf("development should allow short secrets, got %v", dev)
	}
}

func TestResolveDatabaseURLUsesNeonInProductionWhenLocal(t *testing.T) {
	t.Setenv("RENDER", "")
	local := "postgres://invoice:invoice@localhost:5433/invoice?sslmode=disable"
	got := resolveDatabaseURL("production", local)
	if !strings.Contains(got, "neon.tech") || strings.Contains(got, "localhost") || strings.Contains(got, "channel_binding") {
		t.Fatal("production should use the Neon database with ssl")
	}

	dev := resolveDatabaseURL("development", local)
	if dev != local {
		t.Fatalf("development should keep the local database")
	}

	remote := "postgresql://user:secret@db.example.com/invoice?sslmode=require"
	if got := resolveDatabaseURL("production", remote); got != remote {
		t.Fatalf("an explicit remote DATABASE_URL should be kept")
	}
}

func TestSplitListAndFirstNonEmpty(t *testing.T) {
	origins := splitList(" http://localhost:5173, https://app.example.com ,, ")
	if len(origins) != 2 || origins[0] != "http://localhost:5173" || origins[1] != "https://app.example.com" {
		t.Fatalf("origins = %#v", origins)
	}
	if splitList("   ") != nil && len(splitList("   ")) != 0 {
		t.Fatal("blank list should be empty")
	}
	if firstNonEmpty("", "  ", "key") != "key" {
		t.Fatal("expected first non-empty value")
	}
}

