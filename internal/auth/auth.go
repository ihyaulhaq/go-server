package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {

	claims := &jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWt(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(tokenSecret), nil

		},
	)

	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, jwt.ErrSignatureInvalid
	}

	userid, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}

	return userid, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("auth headers missing")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", errors.New("auth headers must start with Bearer")
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))

	if token == "" {
		return "", errors.New("Bearer token missing")
	}

	return token, nil
}

func MakeRefreshToken() (string, error) {

	key := make([]byte, 32)

	_, err := rand.Read(key)

	if err != nil {
		return "", errors.New("Cant make random key")
	}

	encodedKey := hex.EncodeToString(key)

	return encodedKey, nil
}

func GetApiKey(headers http.Header) (string, error) {
	apiHeader := headers.Get("Authorization")

	if apiHeader == "" {
		return "", errors.New("auth headers missing")
	}

	const prefix = "ApiKey "
	if !strings.HasPrefix(apiHeader, prefix) {
		return "", errors.New("auth headers must start with ApiKey")
	}

	token := strings.TrimSpace(strings.TrimPrefix(apiHeader, prefix))

	if token == "" {
		return "", errors.New("ApiKey token missing")
	}

	return token, nil

}
