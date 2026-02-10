package api

import (
	"context"
	"net/http"

	"github.com/ihyaulhaq/go-server/internal/auth"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func (cfg *ApiConfig) MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		userId, err := auth.ValidateJWt(tokenStr, cfg.SecretKey)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
