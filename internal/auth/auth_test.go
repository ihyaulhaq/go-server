package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "super-secret-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hash == "" {
		t.Fatal("expected hash, got empty string")
	}

	if hash == password {
		t.Fatal("hash should not equal password")
	}
}

func TestCheckPasswordHash_Success(t *testing.T) {
	password := "super-secret-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}

	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("compare error: %v", err)
	}

	if !match {
		t.Fatal("expected password to match hash")
	}
}

func TestCheckPasswordHash_Failure(t *testing.T) {
	password := "super-secret-password"
	wrongPassword := "wrong-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}

	match, err := CheckPasswordHash(wrongPassword, hash)
	if err != nil {
		t.Fatalf("compare error: %v", err)
	}

	if match {
		t.Fatal("expected password NOT to match hash")
	}
}

func TestJWT_Success(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()
	expiresIn := time.Minute

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	gotUserID, err := ValidateJWt(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT error: %v", err)
	}

	if gotUserID != userID {
		t.Fatalf("expected userID %v, got %v", userID, gotUserID)
	}
}

func TestJWT_WrongSecret(t *testing.T) {
	userID := uuid.New()

	token, err := MakeJWT(userID, "correct-secret", time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	_, err = ValidateJWt(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestJWT_Expired(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userID, secret, -time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	_, err = ValidateJWt(token, secret)
	if err == nil {
		t.Fatal("expected error for expired token")
	}

	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestJWT_InvalidToken(t *testing.T) {
	_, err := ValidateJWt("not.a.jwt", "secret")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
