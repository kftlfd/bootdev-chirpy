package server

import (
	"chirpy/internal/auth"
	"net/http"
)

func (s *Server) protected(
	mux *http.ServeMux,
	pattern string,
	handler http.HandlerFunc,
) {
	mux.Handle(pattern, auth.WithAuth(s.cfg, s.logger, http.HandlerFunc(handler)))
}

func (s *Server) registerAdminRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /metrics", s.handlerAdmin.MetricsHandler)
	mux.HandleFunc("POST /reset", s.handlerAdmin.ResetMetricsHandler)
}

func (s *Server) registerApiRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", s.handlerAdmin.HandleHealth)

	// users

	mux.HandleFunc("POST /users", s.handlerUsers.CreateUser)
	mux.HandleFunc("POST /login", s.handlerUsers.Login)
	mux.HandleFunc("POST /refresh", s.handlerUsers.RefreshToken)
	mux.HandleFunc("POST /revoke", s.handlerUsers.RevokeToken)
	mux.HandleFunc("POST /polka/webhooks", s.handlerUsers.WebhookUpgradeUser)

	s.protected(mux, "PUT /users", s.handlerUsers.UpdateUser)

	// chirps

	mux.HandleFunc("GET /chirps", s.handlerChirps.GetAllChirps)
	mux.HandleFunc("GET /chirps/{id}", s.handlerChirps.GetChirp)

	s.protected(mux, "POST /chirps", s.handlerChirps.CreateChirp)
	s.protected(mux, "DELETE /chirps/{id}", s.handlerChirps.DeleteChirp)
}

func (s *Server) registerRoutes() {
	admin := http.NewServeMux()
	s.registerAdminRoutes(admin)
	s.mux.Handle("/admin/", http.StripPrefix("/admin", admin))

	api := http.NewServeMux()
	s.registerApiRoutes(api)
	s.mux.Handle("/api/", http.StripPrefix("/api", api))

	s.mux.Handle("/app/", http.StripPrefix("/app", s.handlerApp.ServeAppFiles()))
}
