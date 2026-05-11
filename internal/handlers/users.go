package handlers

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type HandlerUsers struct {
	db *database.Queries
}

func NewHandlerUsers(db *database.Queries) *HandlerUsers {
	if db == nil {
		panic("db is nill")
	}

	return &HandlerUsers{
		db: db,
	}
}

func (h *HandlerUsers) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	reqBody := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	passwordHash, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		log.Printf("Error creating password hash: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := h.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          reqBody.Email,
		HashedPassword: passwordHash,
	})
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
}

func (h *HandlerUsers) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	reqBody := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := h.db.GetUserByEmail(r.Context(), reqBody.Email)
	if err != nil {
		log.Printf("Error getting user: %s", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	passwordOk, err := auth.CheckPasswordHash(reqBody.Password, user.HashedPassword)
	if err != nil {
		log.Printf("Error creating password hash: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !passwordOk {
		w.WriteHeader(http.StatusUnauthorized)
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

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf8")
	w.Write(data)
}
