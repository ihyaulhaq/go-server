package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func replaceBadWords(w string) string {

	badWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	words := strings.Split(w, " ")

	for i, word := range words {
		lowerWord := strings.ToLower(word)

		if badWords[lowerWord] {
			words[i] = "****"

		}
	}
	return strings.Join(words, " ")
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errors struct {
		Error string `json:"error"`
	}
	log.Printf("Responding with error %d: %s", code, msg)

	respondWithJSON(w, code, errors{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Write(data)
}

func (cfg *ApiConfig) ProtectedFunc(
	handler func(http.ResponseWriter, *http.Request),
) http.Handler {
	return cfg.MiddlewareAuth(http.HandlerFunc(handler))
}
