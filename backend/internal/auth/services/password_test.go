package services

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("mysecretpassword", bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == "mysecretpassword" {
		t.Fatal("hash should not equal plaintext")
	}
}

func TestCheckPasswordCorrect(t *testing.T) {
	hash, _ := HashPassword("mysecretpassword", bcrypt.MinCost)
	if err := CheckPassword(hash, "mysecretpassword"); err != nil {
		t.Fatalf("expected correct password to pass: %v", err)
	}
}

func TestCheckPasswordWrong(t *testing.T) {
	hash, _ := HashPassword("mysecretpassword", bcrypt.MinCost)
	if err := CheckPassword(hash, "wrongpassword"); err == nil {
		t.Fatal("expected wrong password to fail")
	}
}
