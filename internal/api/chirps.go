package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ihyaulhaq/go-server/internal/database"
)

type ChirpsResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *ApiConfig) HandleCreateChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 400, "invalid request payload")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	userIDVal := r.Context().Value(userIDContextKey)
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		respondWithError(w, 401, "unauthorized")
		return
	}

	cleanedBody := replaceBadWords(params.Body)
	params.Body = cleanedBody

	chirp, err := cfg.DB.CreateChirps(r.Context(), database.CreateChirpsParams{
		Body:   params.Body,
		UserID: userID,
	})

	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	responseChirp := ChirpsResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	respondWithJSON(w, 201, responseChirp)
}

func (cfg *ApiConfig) HandleGetChirps(w http.ResponseWriter, r *http.Request) {
	authorIdStr := r.URL.Query().Get("author_id")
	sortBy := r.URL.Query().Get("sort")

	var chirps []database.Chirp
	var err error

	if authorIdStr != "" {

		authorId, err := uuid.Parse(authorIdStr)
		if err != nil {
			respondWithError(w, 400, err.Error())
			return
		}

		chirps, err = cfg.DB.GetChirpByAuthor(r.Context(), authorId)
	} else {

		chirps, err = cfg.DB.GetChirps(r.Context())
	}

	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	response := make([]ChirpsResponse, 0, len(chirps))
	for _, c := range chirps {
		response = append(response, ChirpsResponse{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		})
	}

	if strings.ToLower(sortBy) == "desc" {

		sort.Slice(
			response,
			func(i, j int) bool {
				return response[i].CreatedAt.After(response[j].CreatedAt)
			},
		)
	}

	respondWithJSON(w, 200, response)
}

func (cfg *ApiConfig) HandleGetChirp(w http.ResponseWriter, r *http.Request) {

	chirpIdStr := r.PathValue("id")
	chirpId, err := uuid.Parse(chirpIdStr)

	if err != nil {
		respondWithError(w, 400, "invalid chirp id")
		return
	}

	chirp, err := cfg.DB.GetChirp(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, 404, "chirp not found")
		return
	}

	respondWithJSON(w, 200, chirp)
}

func (cfg *ApiConfig) HandleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(userIDContextKey)
	userId, ok := userIDVal.(uuid.UUID)
	if !ok {
		respondWithError(w, 401, "unauthorized")
		return
	}

	chirpIdStr := r.PathValue("id")
	chirpId, err := uuid.Parse(chirpIdStr)
	if err != nil {
		respondWithError(w, 400, "invalid chirp id")
		return
	}

	rows, err := cfg.DB.DeleteChirp(r.Context(), database.DeleteChirpParams{
		ID:     chirpId,
		UserID: userId,
	})

	if err != nil {
		respondWithError(w, 500, "failed to delete chirp")
		return
	}
	if rows == 0 {
		respondWithError(w, 404, "chirp not found or unauthorized")
		return
	}

	respondWithJSON(w, 200, map[string]string{"message": "chirp deleted successfully"})
}
