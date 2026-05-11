package handlers

import (
	"chirpy/internal/auth"
	"chirpy/internal/config"
	"chirpy/internal/database"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type HandlerChirps struct {
	cfg *config.Config
	db  *database.Queries
}

func NewHandlerChirps(cfg *config.Config, db *database.Queries) *HandlerChirps {
	if cfg == nil {
		panic("cfg is nil")
	}
	if db == nil {
		panic("db is nill")
	}

	return &HandlerChirps{
		cfg: cfg,
		db:  db,
	}
}

func createBadWordsMap() map[string]struct{} {
	badWords := map[string]struct{}{}
	addBadWord := func(w string) {
		badWords[w] = struct{}{}
	}

	addBadWord("kerfuffle")
	addBadWord("sharbert")
	addBadWord("fornax")

	return badWords
}

var badWords = createBadWordsMap()

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

func (h *HandlerChirps) CreateChirp(w http.ResponseWriter, r *http.Request) {
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

	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting auth token: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userId, err := auth.ValidateJWT(tokenStr, h.cfg.Env.ServerSecret)
	if err != nil {
		log.Printf("Error validating jwt: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
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

	chirp, err := h.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   censorChirp(reqBody.Body),
		UserID: userId,
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
}

func (h *HandlerChirps) GetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := h.db.GetAllChirps(r.Context())
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
}

func (h *HandlerChirps) GetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/plain; charset=utf8")
		w.Write([]byte("Invalid ID"))
		return
	}

	chirp, err := h.db.GetChirpById(r.Context(), id)
	if err != nil {
		log.Printf("Error getting chirp: %s", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	type ChirpDto struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	data, err := json.Marshal(ChirpDto(chirp))
	if err != nil {
		log.Printf("Error creating response data: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf8")
	w.Write(data)
}
