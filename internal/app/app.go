package app

import (
	"chirpy/internal/config"
	"chirpy/internal/database"
	"chirpy/internal/handlers"
	"chirpy/internal/metrics"
	"chirpy/internal/server"
	"database/sql"
	"net/http"
)

func BuildApp(cfg *config.Config, dbConn *sql.DB) http.Handler {
	db := database.New(dbConn)

	metrics := metrics.New()

	handlerAdmin := handlers.NewHandlerAdmin(cfg, db, metrics)
	handlerUsers := handlers.NewHandlerUsers(cfg, db)
	handlerChirps := handlers.NewHandlerChirps(cfg, db)
	handlerApp := handlers.NewHandlerApp(metrics)

	return server.New(server.Deps{
		HandlerAdmin:  handlerAdmin,
		HandlerUsers:  handlerUsers,
		HandlerChirps: handlerChirps,
		HandlerApp:    handlerApp,
	})
}
