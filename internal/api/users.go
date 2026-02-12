package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ihyaulhaq/go-server/internal/auth"
	"github.com/ihyaulhaq/go-server/internal/database"
)

type UserResponse struct {
	Id          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type UserLoginResponse struct {
	Id           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func (cfg *ApiConfig) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "invalid request payload")
		return
	}

	if params.Email == "" || params.Password == "" {
		respondWithError(w, 400, "email and password are required")
		return
	}

	params.Password, err = auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "something went wrong: cant hash password")
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: params.Password,
	})

	if err != nil {
		respondWithError(w, 500, "something went wrong: cant create user")
		return
	}

	response := UserResponse{
		Id:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	respondWithJSON(w, 201, response)

}

func (cfg *ApiConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "invalid request payload")
		return
	}

	if params.Email == "" || params.Password == "" {
		respondWithError(w, 400, "email and password are required")
		return
	}

	user, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 404, "user not found")
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, 500, "something went wrong: cant check password hash")
		return
	}

	if !match {
		respondWithError(w, 401, "unauthorized: invalid credentials")
		return
	}

	expiresIn := time.Hour
	token, err := auth.MakeJWT(user.ID, cfg.SecretKey, expiresIn)
	if err != nil {
		respondWithError(w, 500, "could not create token")
		return
	}

	const refreshTokenTTL = 24 * time.Hour * 60
	refreshKey, err := auth.MakeRefreshToken()
	expiresAt := time.Now().UTC().Add(refreshTokenTTL)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	refreshToken, err := cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshKey,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		respondWithError(w, 500, "cant create refresh token")
		return
	}

	response := UserLoginResponse{
		Id:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken.Token,
		IsChirpyRed:  user.IsChirpyRed,
	}

	respondWithJSON(w, 200, response)

}

func (cfg *ApiConfig) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	tokenRecord, err := cfg.DB.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	// Check if revoked
	if tokenRecord.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "token revoked")
		return
	}

	// Check expiration
	if time.Now().After(tokenRecord.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "token expired")
		return
	}

	// Create new access token
	accessToken, err := auth.MakeJWT(
		tokenRecord.UserID,
		cfg.SecretKey,
		time.Hour,
	)
	if err != nil {
		respondWithError(w, 500, "could not create access token")
		return
	}

	type response struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, 200, response{
		Token: accessToken,
	})

}

func (cfg *ApiConfig) HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	err = cfg.DB.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, 500, "could not revoke token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *ApiConfig) HandleEditUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "invalid request payload")
		return
	}

	if params.Email == "" || params.Password == "" {
		respondWithError(w, 400, "email and password are required")
		return
	}

	params.Password, err = auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "something went wrong: cant hash password")
		return
	}

	userIDVal := r.Context().Value(userIDContextKey)
	userId, ok := userIDVal.(uuid.UUID)
	if !ok {
		respondWithError(w, 401, "unauthorized")
		return
	}

	newUser, err := cfg.DB.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: params.Password,
		ID:             userId,
	})
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	response := UserResponse{
		Id:          newUser.ID,
		CreatedAt:   newUser.CreatedAt,
		UpdatedAt:   newUser.UpdatedAt,
		Email:       newUser.Email,
		IsChirpyRed: newUser.IsChirpyRed,
	}

	respondWithJSON(w, 200, response)
}

func (cfg *ApiConfig) HandleUpgradeUserToChirpyRed(w http.ResponseWriter, r *http.Request) {
	type upgradeData struct {
		UserID uuid.UUID `json:"user_id"`
	}

	type parameters struct {
		Event string      `json:"event"`
		Data  upgradeData `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	apiKey, err := auth.GetApiKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if apiKey != cfg.PolkaKey {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = cfg.DB.UpgradeUserToChirpyRed(r.Context(), params.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, 404, err.Error())
			return
		}
		respondWithError(w, 500, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
