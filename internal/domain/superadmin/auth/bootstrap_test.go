package auth

import "testing"

func TestAllowBootstrap(t *testing.T) {
	production := &Handler{environment: "production", bootstrapSecret: "bootstrap-secret"}
	if production.allowBootstrap("") || production.allowBootstrap("wrong") {
		t.Fatal("production must reject a missing or wrong secret")
	}
	if !production.allowBootstrap("bootstrap-secret") {
		t.Fatal("production must accept the configured secret")
	}

	lockedProduction := &Handler{environment: "production"}
	if lockedProduction.allowBootstrap("anything") {
		t.Fatal("production without a secret must stay closed")
	}

	development := &Handler{environment: "development"}
	if !development.allowBootstrap("") {
		t.Fatal("development without a secret can create the first superadmin")
	}

	lockedDevelopment := &Handler{environment: "development", bootstrapSecret: "local-secret"}
	if lockedDevelopment.allowBootstrap("wrong") || !lockedDevelopment.allowBootstrap("local-secret") {
		t.Fatal("a configured development secret must be checked")
	}
}
