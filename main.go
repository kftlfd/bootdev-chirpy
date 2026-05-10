package main

import (
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	isDev          bool
	db             *database.Queries
	fileserverHits atomic.Int32
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening DB: %s", err)
	}

	dbQueries := database.New(db)

	cfg := &apiConfig{
		isDev: os.Getenv("PLATFORM") == "dev",
		db:    dbQueries,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Println(err)
		}
	})

	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetMetricsHandler)

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		reqBody := struct {
			Email string `json:"email"`
		}{}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&reqBody); err != nil {
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user, err := cfg.db.CreateUser(r.Context(), reqBody.Email)
		if err != nil {
			log.Printf("Error creating user: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		respBody := struct {
			Id        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Email     string    `json:"email"`
		}{
			Id:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		}
		data, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error creating response body: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		w.Write(data)
	})

	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		reqBody := struct {
			Body   string    `json:"body"`
			UserId uuid.UUID `json:"user_id"`
		}{}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&reqBody); err != nil {
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(reqBody.Body) > 140 {
			type errRespoBody struct {
				Error string `json:"error"`
			}
			respBody := errRespoBody{Error: "Chirp is too long"}
			data, err := json.Marshal(respBody)
			if err != nil {
				log.Printf("Error marshalling JSON: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json; charset=utf8")
			w.Write(data)
			return
		}

		chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
			Body:   censorChirp(reqBody.Body),
			UserID: reqBody.UserId,
		})
		if err != nil {
			log.Printf("Error creating chirp: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		respBody := struct {
			Id        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Body      string    `json:"body"`
			UserId    uuid.UUID `json:"user_id"`
		}{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		}
		data, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error creating response body: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		w.Write(data)
	})

	mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		chirps, err := cfg.db.GetAllChirps(r.Context())
		if err != nil {
			log.Printf("Error getting chirps: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		type ChirpDto struct {
			ID        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Body      string    `json:"body"`
			UserID    uuid.UUID `json:"user_id"`
		}
		dtos := make([]ChirpDto, len(chirps))
		for i, chirp := range chirps {
			dtos[i] = ChirpDto(chirp)
		}
		data, err := json.Marshal(dtos)
		if err != nil {
			log.Printf("Error creating response data: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		w.Write(data)
	})

	mux.Handle("/app/", cfg.middlewareMetricsInc(newAppFsHandler("/app/")))

	server := http.Server{Handler: mux, Addr: ":8080"}
	log.Fatal(server.ListenAndServe())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
	w.Write([]byte(data))
}

func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if !cfg.isDev {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		log.Printf("reset users error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}

func newAppFsHandler(prefix string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.Dir(".")))
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

var badWords = map[string]struct{}{}

func init() {
	addBadWord := func(w string) {
		badWords[w] = struct{}{}
	}
	addBadWord("kerfuffle")
	addBadWord("sharbert")
	addBadWord("fornax")
}

func isBadWord(word string) bool {
	_, ok := badWords[word]
	return ok
}

func censorChirp(chirp string) string {
	words := strings.Split(chirp, " ")

	for i, word := range words {
		if isBadWord(strings.ToLower(word)) {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}
