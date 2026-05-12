package server

import (
	"net/http"
	"time"
)

func (s *Server) withMiddleware() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		s.mux.ServeHTTP(w, r)

		s.logger.Info(
			"http request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	})
}
