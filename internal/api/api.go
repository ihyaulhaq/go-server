package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func HandlerValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		CleanBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 500, "something went wrong")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chrip is too long")
		return
	}

	cleanedResp := replaceBadWords(params.Body)

	respondWithJSON(w, 200, returnVals{
		CleanBody: cleanedResp,
	})
}

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
