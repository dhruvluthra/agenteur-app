package services

import (
	"testing"
)

func TestGenerateRandomTokenUniqueness(t *testing.T) {
	raw1, hash1, err := GenerateRandomToken()
	if err != nil {
		t.Fatal(err)
	}
	raw2, hash2, err := GenerateRandomToken()
	if err != nil {
		t.Fatal(err)
	}
	if raw1 == raw2 {
		t.Fatal("expected unique raw tokens")
	}
	if hash1 == hash2 {
		t.Fatal("expected unique hashes")
	}
}

func TestGenerateRandomTokenLength(t *testing.T) {
	raw, _, err := GenerateRandomToken()
	if err != nil {
		t.Fatal(err)
	}
	// 32 bytes hex-encoded = 64 characters
	if len(raw) != 64 {
		t.Fatalf("expected raw token length 64, got %d", len(raw))
	}
}

func TestHashTokenDeterministic(t *testing.T) {
	raw := "test-token-value"
	h1 := HashToken(raw)
	h2 := HashToken(raw)
	if h1 != h2 {
		t.Fatal("expected deterministic hash")
	}
	if h1 == raw {
		t.Fatal("hash should not equal input")
	}
}
