package utils

import (
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	password := "test-secure-password-123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !valid {
		t.Fatal("expected password verification to succeed")
	}

	valid, err = VerifyPassword("wrong-password", hash)
	if err != nil {
		t.Fatalf("VerifyPassword with wrong password failed: %v", err)
	}
	if valid {
		t.Fatal("expected password verification to fail with wrong password")
	}
}

func TestGenerateAPIKey(t *testing.T) {
	raw, prefix, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey failed: %v", err)
	}

	if len(raw) == 0 {
		t.Fatal("expected non-empty raw key")
	}

	if len(prefix) != 8 {
		t.Fatalf("expected prefix of length 8, got %d", len(prefix))
	}

	if raw[:8] != prefix {
		t.Fatal("prefix should match first 8 chars of raw key")
	}
}
