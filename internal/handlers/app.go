package handlers

import (
	"chirpy/internal/config"
	"net/http"
)

type HandlerApp struct {
	cfg *config.Config
}

func NewHandlerApp(cfg *config.Config) *HandlerApp {
	return &HandlerApp{
		cfg: cfg,
	}
}

func (h *HandlerApp) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (h *HandlerApp) ServeAppFiles(prefix string) http.Handler {
	return h.middlewareMetricsInc(http.StripPrefix(prefix, http.FileServer(http.Dir("."))))
}
