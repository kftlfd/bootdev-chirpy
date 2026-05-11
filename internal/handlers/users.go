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

type userDTO struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func toUserDTO(user database.User) userDTO {
	return userDTO{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
}

func (h *HandlerUsers) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	reqBody := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		u.SendJSON(w, http.StatusInternalServerError, u.D{
			"error": fmt.Sprintf("Error decoding parameters: %s", err),
		})
		return
	}

	passwordHash, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		u.SendJSON(w, http.StatusInternalServerError, u.D{
			"error": fmt.Sprintf("Error creating password hash: %s", err),
		})
		return
	}

	user, err := h.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          reqBody.Email,
		HashedPassword: passwordHash,
	})
	if err != nil {
		u.SendJSON(w, http.StatusInternalServerError, u.D{
			"error": fmt.Sprintf("Error creating user: %s", err),
		})
		return
	}

	u.SendJSON(w, http.StatusCreated, toUserDTO(user))
}

func (h *HandlerUsers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		u.SendJSON(w, http.StatusUnauthorized, u.D{
			"error": fmt.Sprintf("No auth token: %s", err),
		})
		return
	}

	userId, err := auth.ValidateJWT(token, h.cfg.Env.ServerSecret)
	if err != nil {
		u.SendJSON(w, http.StatusUnauthorized, u.D{
			"error": fmt.Sprintf("Invalid token: %s", err),
		})
		return
	}

	reqBody := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err = json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		u.SendJSON(w, http.StatusBadRequest, u.D{
			"error": fmt.Sprintf("Invalid request body: %s", err),
		})
		return
	}

	newHash, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		u.SendJSON(w, http.StatusInternalServerError, u.D{
			"error": fmt.Sprintf("Error hashing password: %s", err),
		})
		return
	}

	user, err := h.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userId,
		Email:          reqBody.Email,
		HashedPassword: newHash,
	})
	if err != nil {
		u.SendJSON(w, http.StatusInternalServerError, u.D{
			"error": fmt.Sprintf("Error updating user: %s", err),
		})
		return
	}

	u.SendJSON(w, http.StatusOK, toUserDTO(user))
}

func (h *HandlerUsers) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	reqBody := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		u.SendJSON(w, http.StatusInternalServerError, u.D{
			"error": fmt.Sprintf("Error decoding params: %s", err),
		})
		return
	}

	user, err := h.db.GetUserByEmail(r.Context(), reqBody.Email)
	if err != nil {
		u.SendJSON(w, http.StatusNotFound, u.D{
			"error": fmt.Sprintf("Error getting user: %s", err),
		})
		return
	}

	passwordOk, err := auth.CheckPasswordHash(reqBody.Password, user.HashedPassword)
	if err != nil {
		u.SendJSON(w, http.StatusInternalServerError, u.D{
			"error": fmt.Sprintf("Error creating password hash: %s", err),
		})
		return
	}
	if !passwordOk {
		u.SendJSON(w, http.StatusUnauthorized, u.D{
			"error": "Invalid password",
		})
		return
	}

	token, err := auth.MakeJWT(user.ID, h.cfg.Env.ServerSecret, time.Hour)
	if err != nil {
		u.SendJSON(w, http.StatusInternalServerError, u.D{
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
		u.SendJSON(w, http.StatusInternalServerError, u.D{
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
		u.SendJSON(w, http.StatusUnauthorized, u.D{
			"error": fmt.Sprintf("Error getting auth token: %s", err),
		})
		return
	}

	token, err := h.db.GetRefreshToken(r.Context(), rToken)
	if err != nil {
		u.SendJSON(w, http.StatusUnauthorized, u.D{
			"error": fmt.Sprintf("Error getting refresh token from DB: %s", err),
		})
		return
	}

	jwt, err := auth.MakeJWT(token.UserID, h.cfg.Env.ServerSecret, time.Hour)
	if err != nil {
		u.SendJSON(w, http.StatusInternalServerError, u.D{
			"error": fmt.Sprintf("Error creating jwt: %s", err),
		})
		return
	}

	u.SendJSON(w, http.StatusOK, u.D{
		"token": jwt,
	})
}

func (h *HandlerUsers) RevokeToken(w http.ResponseWriter, r *http.Request) {
	rToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		u.SendJSON(w, http.StatusBadRequest, u.D{
			"error": fmt.Sprintf("Error getting auth token: %s", err),
		})
		return
	}

	err = h.db.MarkTokenRevoked(r.Context(), rToken)
	if err != nil {
		u.SendJSON(w, http.StatusBadRequest, u.D{
			"error": fmt.Sprintf("Error marking token revoked: %s", err),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
