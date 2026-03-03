package auth

import (
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	token, expiresAt, err := GenerateToken("alice", "admin", "test-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if token == "" {
		t.Error("token should not be empty")
	}
	if expiresAt == 0 {
		t.Error("expiresAt should not be zero")
	}
	// expiresAt should be roughly 1 hour from now
	now := time.Now().Unix()
	diff := expiresAt - now
	if diff < 3590 || diff > 3610 {
		t.Errorf("expiresAt drift: %d seconds from now, expected ~3600", diff)
	}
}

func TestValidateToken(t *testing.T) {
	secret := "my-secret"
	token, _, err := GenerateToken("bob", "viewer", secret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.Username != "bob" {
		t.Errorf("Username = %q, want %q", claims.Username, "bob")
	}
	if claims.Role != "viewer" {
		t.Errorf("Role = %q, want %q", claims.Role, "viewer")
	}
	if claims.Issuer != "infrawatch" {
		t.Errorf("Issuer = %q, want %q", claims.Issuer, "infrawatch")
	}
}

func TestValidateTokenWrongSecret(t *testing.T) {
	token, _, err := GenerateToken("charlie", "admin", "correct-secret", time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	_, err = ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Error("Expected error when validating with wrong secret")
	}
}

func TestValidateTokenMalformed(t *testing.T) {
	_, err := ValidateToken("not.a.valid.jwt.token", "secret")
	if err == nil {
		t.Error("Expected error for malformed token")
	}
}

func TestValidateTokenEmpty(t *testing.T) {
	_, err := ValidateToken("", "secret")
	if err == nil {
		t.Error("Expected error for empty token string")
	}
}

func TestTokenExpiry(t *testing.T) {
	// Generate a token that is already expired
	token, _, err := GenerateToken("dave", "admin", "secret", -time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	_, err = ValidateToken(token, "secret")
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

func TestMultipleTokensUnique(t *testing.T) {
	t1, _, _ := GenerateToken("user1", "admin", "secret", time.Hour)
	t2, _, _ := GenerateToken("user2", "viewer", "secret", time.Hour)
	if t1 == t2 {
		t.Error("Different users should produce different tokens")
	}
}
