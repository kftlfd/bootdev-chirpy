package server

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/healthz", s.handlerAdmin.HandleHealth)
	s.mux.HandleFunc("GET /admin/metrics", s.handlerAdmin.MetricsHandler)
	s.mux.HandleFunc("POST /admin/reset", s.handlerAdmin.ResetMetricsHandler)

	s.mux.HandleFunc("POST /api/users", s.handlerUsers.CreateUser)
	s.mux.Handle("PUT /api/users", s.handlerUsers.UpdateUser())
	s.mux.HandleFunc("POST /api/login", s.handlerUsers.Login)
	s.mux.HandleFunc("POST /api/refresh", s.handlerUsers.RefreshToken)
	s.mux.HandleFunc("POST /api/revoke", s.handlerUsers.RevokeToken)

	s.mux.Handle("POST /api/chirps", s.handlerChirps.CreateChirp())
	s.mux.HandleFunc("GET /api/chirps", s.handlerChirps.GetAllChirps)
	s.mux.HandleFunc("GET /api/chirps/{id}", s.handlerChirps.GetChirp)
	s.mux.Handle("DELETE /api/chirps/{id}", s.handlerChirps.DeleteChirp())

	s.mux.Handle("/app/", s.handlerApp.ServeAppFiles("/app/"))
}
