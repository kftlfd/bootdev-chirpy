package handlers

import (
	"chirpy/internal/config"
	"chirpy/internal/metrics"
	"net/http"
)

type HandlerApp struct {
	cfg     *config.Config
	metrics *metrics.Metrics
}

func NewHandlerApp(cfg *config.Config, m *metrics.Metrics) *HandlerApp {
	if cfg == nil {
		panic("config is nil")
	}
	if m == nil {
		panic("metrics is nil")
	}

	return &HandlerApp{
		cfg:     cfg,
		metrics: m,
	}
}

// @summary	Static file
// @tags		App
// @router		/app/{path} [get]
// @param		path	path	string	false	"path to file"
func (h *HandlerApp) ServeAppFiles() http.Handler {
	fs := http.FileServer(http.Dir(h.cfg.FsRoot))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.metrics.FileserverHits.Add(1)
		fs.ServeHTTP(w, r)
	})
}
