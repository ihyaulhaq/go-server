package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/ihyaulhaq/go-server/internal/api"
	"github.com/ihyaulhaq/go-server/internal/database"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading env file")
	}

	const filepathRoot = "."
	const port = "8080"

	dbURL := os.Getenv("DB_URL")
	enviroment := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	jwtKey := os.Getenv("SECRET")
	polka_key := os.Getenv("POLKA_KEY")

	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	apiCfg := api.ApiConfig{
		FileserverHits: atomic.Int32{},
		DB:             dbQueries,
		Platform:       enviroment,
		SecretKey:      jwtKey,
		PolkaKey:       polka_key,
	}

	fsHandler := apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))

	mux := http.NewServeMux()

	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /api/healthz", api.HandlerReadiness)

	mux.HandleFunc("GET /admin/metrics", apiCfg.HandlerMetrics)

	mux.HandleFunc("POST /admin/reset", apiCfg.HandlerReset)

	mux.HandleFunc("POST /api/users", apiCfg.HandleCreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.HandleLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.HandleRefreshToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.HandleRevokeToken)
	mux.Handle(
		"PUT /api/users",
		apiCfg.ProtectedFunc(apiCfg.HandleEditUser),
	)

	mux.Handle("POST /api/chirps",
		apiCfg.ProtectedFunc(apiCfg.HandleCreateChirps),
	)
	mux.HandleFunc("GET /api/chirps", apiCfg.HandleGetChirps)
	mux.HandleFunc("GET /api/chirps/{id}", apiCfg.HandleGetChirp)
	mux.Handle(
		"DELETE /api/chirps/{id}",
		apiCfg.ProtectedFunc(apiCfg.HandleDeleteChirp),
	)

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.HandleUpgradeUserToChirpyRed)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
