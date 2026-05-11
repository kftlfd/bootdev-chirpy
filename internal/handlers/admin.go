package handlers

import (
	"chirpy/internal/config"
	"chirpy/internal/database"
	"fmt"
	"log"
	"net/http"
)

type HandlerAdmin struct {
	cfg *config.Config
	db  *database.Queries
}

func NewHandlerAdmin(cfg *config.Config, db *database.Queries) *HandlerAdmin {
	if cfg == nil {
		panic("cfg is nil")
	}
	if db == nil {
		panic("db is nil")
	}

	return &HandlerAdmin{
		cfg: cfg,
		db:  db,
	}
}

func (h *HandlerAdmin) HandleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Println(err)
	}
}

func (h *HandlerAdmin) MetricsHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, h.cfg.FileserverHits.Load())
	w.Write([]byte(data))
}

func (h *HandlerAdmin) ResetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Env.IsDev {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err := h.db.ResetUsers(r.Context())
	if err != nil {
		log.Printf("reset users error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	h.cfg.FileserverHits.Store(0)
}
