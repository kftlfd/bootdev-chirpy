package handlers

import (
	"chirpy/internal/metrics"
	"net/http"
)

type HandlerApp struct {
	metrics *metrics.Metrics
}

func NewHandlerApp(m *metrics.Metrics) *HandlerApp {
	if m == nil {
		panic("metrics is nil")
	}

	return &HandlerApp{
		metrics: m,
	}
}

func (h *HandlerApp) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.metrics.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (h *HandlerApp) ServeAppFiles(prefix string) http.Handler {
	return h.middlewareMetricsInc(http.StripPrefix(prefix, http.FileServer(http.Dir("public"))))
}
