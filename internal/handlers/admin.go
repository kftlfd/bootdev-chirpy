package handlers

import (
	"chirpy/internal/config"
	"chirpy/internal/database"
	"chirpy/internal/httpx"
	"chirpy/internal/metrics"
	"fmt"
	"log/slog"
	"net/http"
)

type HandlerAdmin struct {
	cfg     *config.Config
	db      *database.Queries
	metrics *metrics.Metrics
	res     *httpx.Responder
}

func NewHandlerAdmin(cfg *config.Config, logger *slog.Logger, db *database.Queries, m *metrics.Metrics) *HandlerAdmin {
	if logger == nil {
		panic("logger is nil")
	}
	if cfg == nil {
		panic("cfg is nil")
	}
	if db == nil {
		panic("db is nil")
	}
	if m == nil {
		panic("metrics is nil")
	}

	log := logger.With("module", "handler-admin")

	return &HandlerAdmin{
		cfg:     cfg,
		db:      db,
		metrics: m,
		res:     httpx.NewResponder(log),
	}
}

// @summary	Health
// @tags		API admin
// @router		/api/healthz [get]
// @success	200	{string}	string	Ok
func (h *HandlerAdmin) HandleHealth(w http.ResponseWriter, _ *http.Request) {
	h.res.Text(w, http.StatusOK, []byte("OK"))
}

func getMetricsHtml(visits int32) string {
	return fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, visits)
}

// @summary	View Metrics
// @tags		API admin
// @router		/admin/metrics [get]
// @produce	html
func (h *HandlerAdmin) MetricsHandler(w http.ResponseWriter, _ *http.Request) {
	html := getMetricsHtml(h.metrics.FileserverHits.Load())
	h.res.HTML(w, http.StatusOK, []byte(html))
}

// @summary	Reset DB and metrics
// @tags		API admin
// @router		/admin/reset [post]
// @Produce	json
func (h *HandlerAdmin) ResetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.IsDev {
		h.res.JSONError(w, http.StatusForbidden, "unsafe operation")
		return
	}

	err := h.db.ResetUsers(r.Context())
	if err != nil {
		h.res.JSONError(w, http.StatusInternalServerError,
			fmt.Sprintf("reset users error: %s", err))
		return
	}

	h.metrics.FileserverHits.Store(0)
	h.res.JSON(w, http.StatusOK, "OK")
}
