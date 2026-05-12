package handlers

import (
	"chirpy/internal/config"
	"chirpy/internal/database"
	"chirpy/internal/metrics"
	u "chirpy/internal/utils"
	"fmt"
	"log"
	"net/http"
)

type HandlerAdmin struct {
	cfg     *config.Config
	db      *database.Queries
	metrics *metrics.Metrics
}

func NewHandlerAdmin(cfg *config.Config, db *database.Queries, m *metrics.Metrics) *HandlerAdmin {
	if cfg == nil {
		panic("cfg is nil")
	}
	if db == nil {
		panic("db is nil")
	}
	if m == nil {
		panic("metrics is nil")
	}

	return &HandlerAdmin{
		cfg:     cfg,
		db:      db,
		metrics: m,
	}
}

func (h *HandlerAdmin) HandleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Printf("write: %v", err)
	}
}

func getMetricsHtml(visits int32) string {
	return fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, visits)
}

func (h *HandlerAdmin) MetricsHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := getMetricsHtml(h.metrics.FileserverHits.Load())
	if _, err := w.Write([]byte(html)); err != nil {
		log.Printf("write: %v", err)
	}
}

func (h *HandlerAdmin) ResetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.IsDev {
		u.SendJSONError(w, http.StatusForbidden, "unsafe operation")
		return
	}

	err := h.db.ResetUsers(r.Context())
	if err != nil {
		u.SendJSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("reset users error: %s", err))
		return
	}

	h.metrics.FileserverHits.Store(0)
	u.SendJSON(w, http.StatusOK, "OK")
}
