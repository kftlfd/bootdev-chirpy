package handlers

import (
	"chirpy/internal/auth"
	"chirpy/internal/config"
	"chirpy/internal/database"
	u "chirpy/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type HandlerUsers struct {
	cfg *config.Config
	db  *database.Queries
}

func NewHandlerUsers(cfg *config.Config, db *database.Queries) *HandlerUsers {
	if cfg == nil {
		panic("cfg is nil")
	}
	if db == nil {
		panic("db is nill")
	}

	return &HandlerUsers{
		cfg: cfg,
		db:  db,
	}
}

func (h *HandlerUsers) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	reqBody := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		u.SendJSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error decoding parameters: %s", err),
		})
		return
	}

	passwordHash, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		u.SendJSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error creating password hash: %s", err),
		})
		return
	}

	user, err := h.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          reqBody.Email,
		HashedPassword: passwordHash,
	})
	if err != nil {
		u.SendJSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error creating user: %s", err),
		})
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
	u.SendJSON(w, http.StatusCreated, respBody)
}

func (h *HandlerUsers) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	reqBody := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		u.SendJSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error decoding params: %s", err),
		})
		return
	}

	user, err := h.db.GetUserByEmail(r.Context(), reqBody.Email)
	if err != nil {
		u.SendJSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("Error getting user: %s", err),
		})
		return
	}

	passwordOk, err := auth.CheckPasswordHash(reqBody.Password, user.HashedPassword)
	if err != nil {
		u.SendJSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error creating password hash: %s", err),
		})
		return
	}
	if !passwordOk {
		u.SendJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "Invalid password",
		})
		return
	}

	token, err := auth.MakeJWT(user.ID, h.cfg.Env.ServerSecret, time.Hour)
	if err != nil {
		u.SendJSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error creating jwt: %s", err),
		})
		return
	}

	rtData := auth.MakeRefreshToken()
	refreshToken, err := h.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     rtData,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		u.SendJSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error creating refresh token: %s", err),
		})
		return
	}

	respBody := struct {
		Id           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}{
		Id:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken.Token,
	}
	u.SendJSON(w, http.StatusOK, respBody)
}

func (h *HandlerUsers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	rToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		u.SendJSON(w, http.StatusUnauthorized, map[string]string{
			"error": fmt.Sprintf("Error getting auth token: %s", err),
		})
		return
	}

	token, err := h.db.GetRefreshToken(r.Context(), rToken)
	if err != nil {
		u.SendJSON(w, http.StatusUnauthorized, map[string]string{
			"error": fmt.Sprintf("Error getting refresh token from DB: %s", err),
		})
		return
	}

	jwt, err := auth.MakeJWT(token.UserID, h.cfg.Env.ServerSecret, time.Hour)
	if err != nil {
		u.SendJSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Error creating jwt: %s", err),
		})
		return
	}

	u.SendJSON(w, http.StatusOK, map[string]string{
		"token": jwt,
	})
}

func (h *HandlerUsers) RevokeToken(w http.ResponseWriter, r *http.Request) {
	rToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		u.SendJSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Error getting auth token: %s", err),
		})
		return
	}

	err = h.db.MarkTokenRevoked(r.Context(), rToken)
	if err != nil {
		u.SendJSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Error marking token revoked: %s", err),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
