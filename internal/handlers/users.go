package handlers

import (
	"chirpy/internal/auth"
	"chirpy/internal/config"
	"chirpy/internal/database"
	u "chirpy/internal/utils"
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
	Id          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func toUserDTO(user database.User) userDTO {
	return userDTO{
		Id:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}
}

type userWithTokensDTO struct {
	userDTO
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func toUserWithTokensDTO(user database.User, jwt string, refreshToken string) userWithTokensDTO {
	return userWithTokensDTO{
		userDTO:      toUserDTO(user),
		Token:        jwt,
		RefreshToken: refreshToken,
	}
}

func (h *HandlerUsers) CreateUser(w http.ResponseWriter, r *http.Request) {
	reqBody := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	if err := u.DecodeJSON(r, &reqBody); err != nil {
		u.SendJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error decoding parameters: %s", err))
		return
	}

	passwordHash, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		u.SendJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error creating password hash: %s", err))
		return
	}

	user, err := h.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          reqBody.Email,
		HashedPassword: passwordHash,
	})
	if err != nil {
		u.SendJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error creating user: %s", err))
		return
	}

	u.SendJSON(w, http.StatusCreated, toUserDTO(user))
}

func (h *HandlerUsers) UpdateUser() http.Handler {
	return auth.WithAuth(h.cfg, http.HandlerFunc(h.updateUser))
}

func (h *HandlerUsers) updateUser(w http.ResponseWriter, r *http.Request) {
	reqBody := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	if err := u.DecodeJSON(r, &reqBody); err != nil {
		u.SendJSONError(w, http.StatusBadRequest,
			fmt.Sprintf("Invalid request body: %s", err))
		return
	}

	newHash, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		u.SendJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error hashing password: %s", err))
		return
	}

	userId := auth.RequireAuthUser(r.Context())

	user, err := h.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userId,
		Email:          reqBody.Email,
		HashedPassword: newHash,
	})
	if err != nil {
		u.SendJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error updating user: %s", err))
		return
	}

	u.SendJSON(w, http.StatusOK, toUserDTO(user))
}

func (h *HandlerUsers) Login(w http.ResponseWriter, r *http.Request) {
	reqBody := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	if err := u.DecodeJSON(r, &reqBody); err != nil {
		u.SendJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error decoding params: %s", err))
		return
	}

	user, err := h.db.GetUserByEmail(r.Context(), reqBody.Email)
	if err != nil {
		u.SendJSONError(w, http.StatusNotFound,
			fmt.Sprintf("Error getting user: %s", err))
		return
	}

	passwordOk, err := auth.CheckPasswordHash(reqBody.Password, user.HashedPassword)
	if err != nil {
		u.SendJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error creating password hash: %s", err))
		return
	}
	if !passwordOk {
		u.SendJSONError(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	token, err := auth.MakeJWT(user.ID, h.cfg.Env.ServerSecret, time.Hour)
	if err != nil {
		u.SendJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error creating jwt: %s", err))
		return
	}

	refreshToken, err := h.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     auth.MakeRefreshToken(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		u.SendJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error creating refresh token: %s", err))
		return
	}

	respBody := toUserWithTokensDTO(user, token, refreshToken.Token)

	u.SendJSON(w, http.StatusOK, respBody)
}

func (h *HandlerUsers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	rToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		u.SendJSONError(w, http.StatusUnauthorized,
			fmt.Sprintf("Error getting auth token: %s", err))
		return
	}

	token, err := h.db.GetRefreshToken(r.Context(), rToken)
	if err != nil {
		u.SendJSONError(w, http.StatusUnauthorized,
			fmt.Sprintf("Error getting refresh token from DB: %s", err))
		return
	}

	jwt, err := auth.MakeJWT(token.UserID, h.cfg.Env.ServerSecret, time.Hour)
	if err != nil {
		u.SendJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error creating jwt: %s", err))
		return
	}

	u.SendJSON(w, http.StatusOK, u.D{
		"token": jwt,
	})
}

func (h *HandlerUsers) RevokeToken(w http.ResponseWriter, r *http.Request) {
	rToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		u.SendJSONError(w, http.StatusBadRequest,
			fmt.Sprintf("Error getting auth token: %s", err))
		return
	}

	err = h.db.MarkTokenRevoked(r.Context(), rToken)
	if err != nil {
		u.SendJSONError(w, http.StatusBadRequest,
			fmt.Sprintf("Error marking token revoked: %s", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *HandlerUsers) WebhookUpgradeUser(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != h.cfg.Env.PolkaKey {
		u.SendJSONError(w, http.StatusUnauthorized, "Invalid API key")
		return
	}

	reqBody := struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}{}

	if err := u.DecodeJSON(r, &reqBody); err != nil {
		u.SendJSONError(w, http.StatusBadRequest,
			fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if reqBody.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userId, err := uuid.Parse(reqBody.Data.UserId)
	if err != nil {
		u.SendJSONError(w, http.StatusBadRequest,
			fmt.Sprintf("invalid user ID: %v", err))
		return
	}

	_, err = h.db.SetUserChirpyRedStatus(r.Context(), database.SetUserChirpyRedStatusParams{
		ID:          userId,
		IsChirpyRed: true,
	})
	if err != nil {
		u.SendJSONError(w, http.StatusNotFound,
			fmt.Sprintf("user not found: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
