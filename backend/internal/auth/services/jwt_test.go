package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

const testSecret = "test-secret-key-for-testing-only!"

func TestJWTGenerateAndValidate(t *testing.T) {
	uid := uuid.New()
	token, err := GenerateAccessToken(uid, "test@example.com", false, testSecret, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := ValidateAccessToken(token, testSecret)
	if err != nil {
		t.Fatalf("expected valid token: %v", err)
	}
	if claims.UserID != uid {
		t.Fatalf("expected uid %s, got %s", uid, claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", claims.Email)
	}
	if claims.IsSuperadmin {
		t.Fatal("expected IsSuperadmin=false")
	}
}

func TestJWTExpiredToken(t *testing.T) {
	uid := uuid.New()
	token, err := GenerateAccessToken(uid, "test@example.com", false, testSecret, -1*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ValidateAccessToken(token, testSecret)
	if err == nil {
		t.Fatal("expected expired token to fail validation")
	}
}

func TestJWTWrongSecret(t *testing.T) {
	uid := uuid.New()
	token, err := GenerateAccessToken(uid, "test@example.com", false, testSecret, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ValidateAccessToken(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected wrong secret to fail validation")
	}
}

func TestJWTParseUnvalidatedExpired(t *testing.T) {
	uid := uuid.New()
	token, err := GenerateAccessToken(uid, "test@example.com", false, testSecret, -1*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := ParseAccessTokenUnvalidated(token, testSecret)
	if err != nil {
		t.Fatalf("expected unvalidated parse to succeed for expired token: %v", err)
	}
	if claims.UserID != uid {
		t.Fatalf("expected uid %s, got %s", uid, claims.UserID)
	}
}

func TestJWTParseUnvalidatedWrongSecret(t *testing.T) {
	uid := uuid.New()
	token, err := GenerateAccessToken(uid, "test@example.com", false, testSecret, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ParseAccessTokenUnvalidated(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected wrong secret to fail even unvalidated parse")
	}
}
