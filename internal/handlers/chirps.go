package handlers

import (
	"chirpy/internal/auth"
	"chirpy/internal/config"
	"chirpy/internal/database"
	"chirpy/internal/httpx"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type HandlerChirps struct {
	cfg    *config.Config
	logger *slog.Logger
	db     *database.Queries
	res    *httpx.Responder
}

func NewHandlerChirps(cfg *config.Config, logger *slog.Logger, db *database.Queries) *HandlerChirps {
	if logger == nil {
		panic("logger is nil")
	}
	if cfg == nil {
		panic("cfg is nil")
	}
	if db == nil {
		panic("db is nill")
	}

	log := logger.With("module", "handler-chirps")

	return &HandlerChirps{
		cfg:    cfg,
		logger: log,
		db:     db,
		res:    httpx.NewResponder(log),
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
	return auth.WithAuth(h.cfg, h.logger, http.HandlerFunc(h.createChirp))
}

func (h *HandlerChirps) createChirp(w http.ResponseWriter, r *http.Request) {
	reqBody := struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}{}

	if err := httpx.DecodeJSON(r, &reqBody); err != nil {
		h.res.JSONError(w, http.StatusBadRequest,
			fmt.Sprintf("Error decoding parameters: %s", err))
		return
	}

	if len(reqBody.Body) > 140 {
		h.res.JSONError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	userId := auth.RequireAuthUser(r.Context())

	chirp, err := h.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   censorChirp(reqBody.Body),
		UserID: userId,
	})
	if err != nil {
		h.res.JSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error creating chirp: %s", err))
		return
	}

	h.res.JSON(w, http.StatusCreated, toChirpDTO(chirp))
}

func (h *HandlerChirps) GetAllChirps(w http.ResponseWriter, r *http.Request) {
	authorId := r.URL.Query().Get("author_id")
	sortOrd := r.URL.Query().Get("sort")

	filter := database.GetAllChirpsParams{}

	if len(authorId) > 0 {
		userId, err := uuid.Parse(authorId)
		if err != nil {
			h.res.JSONError(w, http.StatusBadRequest,
				fmt.Sprintf("invalid author ID: %v", err))
			return
		}
		filter.HasUserID = true
		filter.UserID = userId
	}

	chirps, err := h.db.GetAllChirps(r.Context(), filter)
	if err != nil {
		h.res.JSONError(w, http.StatusInternalServerError,
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

	h.res.JSON(w, http.StatusOK, dtos)
}

func (h *HandlerChirps) GetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.res.JSONError(w, http.StatusBadRequest,
			fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	chirp, err := h.db.GetChirpById(r.Context(), id)
	if err != nil {
		h.res.JSONError(w, http.StatusNotFound,
			fmt.Sprintf("Error getting chirp: %s", err))
		return
	}

	h.res.JSON(w, http.StatusOK, toChirpDTO(chirp))
}

func (h *HandlerChirps) DeleteChirp() http.Handler {
	return auth.WithAuth(h.cfg, h.logger, http.HandlerFunc(h.deleteChirp))
}

func (h *HandlerChirps) deleteChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.res.JSONError(w, http.StatusBadRequest,
			fmt.Sprintf("Invalid url id: %s", err))
		return
	}

	chirp, err := h.db.GetChirpById(r.Context(), id)
	if err != nil {
		h.res.JSONError(w, http.StatusNotFound,
			fmt.Sprintf("not found: %s", err))
		return
	}

	userId := auth.RequireAuthUser(r.Context())

	if chirp.UserID != userId {
		h.res.JSONError(w, http.StatusForbidden, "can delete only your own chirps")
		return
	}

	err = h.db.DeleteChirp(r.Context(), database.DeleteChirpParams{
		UserID: userId,
		ID:     id,
	})
	if err != nil {
		h.res.JSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error deleting chirp: %s", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
