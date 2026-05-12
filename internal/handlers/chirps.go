package handlers

import (
	"chirpy/internal/auth"
	"chirpy/internal/config"
	"chirpy/internal/database"
	u "chirpy/internal/utils"
	"fmt"
	"net/http"
	"sort"
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

type chirpDto struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func toChirpDTO(chirp database.Chirp) chirpDto {
	return chirpDto(chirp)
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

func (h *HandlerChirps) CreateChirp() http.Handler {
	return auth.WithAuth(h.cfg, http.HandlerFunc(h.createChirp))
}

func (h *HandlerChirps) createChirp(w http.ResponseWriter, r *http.Request) {
	reqBody := struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}{}

	if err := u.DecodeJSON(r, &reqBody); err != nil {
		u.SendJSONError(w, http.StatusBadRequest,
			fmt.Sprintf("Error decoding parameters: %s", err))
		return
	}

	if len(reqBody.Body) > 140 {
		u.SendJSONError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	userId := auth.RequireAuthUser(r.Context())

	chirp, err := h.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   censorChirp(reqBody.Body),
		UserID: userId,
	})
	if err != nil {
		u.SendJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error creating chirp: %s", err))
		return
	}

	u.SendJSON(w, http.StatusCreated, toChirpDTO(chirp))
}

func (h *HandlerChirps) GetAllChirps(w http.ResponseWriter, r *http.Request) {
	authorId := r.URL.Query().Get("author_id")
	sortOrd := r.URL.Query().Get("sort")

	filter := database.GetAllChirpsParams{}

	if len(authorId) > 0 {
		userId, err := uuid.Parse(authorId)
		if err != nil {
			u.SendJSONError(w, http.StatusBadRequest,
				fmt.Sprintf("invalid author ID: %v", err))
			return
		}
		filter.HasUserID = true
		filter.UserID = userId
	}

	chirps, err := h.db.GetAllChirps(r.Context(), filter)
	if err != nil {
		u.SendJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error getting chirps: %s", err))
		return
	}

	if sortOrd == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	}

	dtos := make([]chirpDto, len(chirps))
	for i, chirp := range chirps {
		dtos[i] = toChirpDTO(chirp)
	}

	u.SendJSON(w, http.StatusOK, dtos)
}

func (h *HandlerChirps) GetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		u.SendJSONError(w, http.StatusBadRequest,
			fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	chirp, err := h.db.GetChirpById(r.Context(), id)
	if err != nil {
		u.SendJSONError(w, http.StatusNotFound,
			fmt.Sprintf("Error getting chirp: %s", err))
		return
	}

	u.SendJSON(w, http.StatusOK, toChirpDTO(chirp))
}

func (h *HandlerChirps) DeleteChirp() http.Handler {
	return auth.WithAuth(h.cfg, http.HandlerFunc(h.deleteChirp))
}

func (h *HandlerChirps) deleteChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		u.SendJSONError(w, http.StatusBadRequest,
			fmt.Sprintf("Invalid url id: %s", err))
		return
	}

	chirp, err := h.db.GetChirpById(r.Context(), id)
	if err != nil {
		u.SendJSONError(w, http.StatusNotFound,
			fmt.Sprintf("not found: %s", err))
		return
	}

	userId := auth.RequireAuthUser(r.Context())

	if chirp.UserID != userId {
		u.SendJSONError(w, http.StatusForbidden, "can delete only your own chirps")
		return
	}

	err = h.db.DeleteChirp(r.Context(), database.DeleteChirpParams{
		UserID: userId,
		ID:     id,
	})
	if err != nil {
		u.SendJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error deleting chirp: %s", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
