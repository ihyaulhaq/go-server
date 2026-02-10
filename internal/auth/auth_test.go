package auth

import "testing"

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
