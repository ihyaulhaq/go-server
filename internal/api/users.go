package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ihyaulhaq/go-server/internal/auth"
	"github.com/ihyaulhaq/go-server/internal/database"
)

type UserResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

type UserLoginResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
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
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, 201, response)

}

func (cfg *ApiConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email        string `json:"email"`
		Password     string `json:"password"`
		ExpiresInSec *int   `json:"expires_in_seconds"`
	}

	expiresIn := time.Hour

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

	if params.ExpiresInSec != nil {
		requested := time.Duration(*params.ExpiresInSec) * time.Second
		if requested < time.Hour && requested > 0 {
			expiresIn = requested
		}
	}

	token, err := auth.MakeJWT(user.ID, cfg.SecretKey, expiresIn)
	if err != nil {
		respondWithError(w, 500, "could not create token")
		return
	}

	response := UserLoginResponse{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	}

	respondWithJSON(w, 200, response)

}
