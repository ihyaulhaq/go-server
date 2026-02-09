package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *ApiConfig) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}
	type returnVals struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), params.Email)

	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 201, returnVals{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})

}

func (cfg *ApiConfig) HandleDeleteUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		respondWithError(w, http.StatusForbidden, "forbidden")
		return
	}

	err := cfg.DB.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
