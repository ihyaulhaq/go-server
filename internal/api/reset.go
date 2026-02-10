package api

import "net/http"

func (cfg *ApiConfig) HandlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		respondWithError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := cfg.DB.DeleteUsers(r.Context()); err != nil {
		respondWithError(w, 500, "something went wrong: cant delete users")
		return
	}

	if err := cfg.DB.DeleteChirps(r.Context()); err != nil {
		respondWithError(w, 500, "something went wrong: cant delete chirps")
		return
	}

	respondWithJSON(w, 200, map[string]string{
		"message": "reset successful",
	})

}
