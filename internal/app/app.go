package app

import (
	"chirpy/internal/config"
	"chirpy/internal/database"
	"chirpy/internal/handlers"
	"chirpy/internal/metrics"
	"chirpy/internal/server"
	"database/sql"
	"log/slog"
	"net/http"
)

func BuildApp(cfg *config.Config, dbConn *sql.DB, logger *slog.Logger) http.Handler {
	db := database.New(dbConn)

	metrics := metrics.New()

	handlerAdmin := handlers.NewHandlerAdmin(cfg, logger, db, metrics)
	handlerUsers := handlers.NewHandlerUsers(cfg, logger, db)
	handlerChirps := handlers.NewHandlerChirps(cfg, logger, db)
	handlerApp := handlers.NewHandlerApp(cfg, metrics)

	return server.New(server.Deps{
		Config:        cfg,
		Logger:        logger,
		HandlerAdmin:  handlerAdmin,
		HandlerUsers:  handlerUsers,
		HandlerChirps: handlerChirps,
		HandlerApp:    handlerApp,
	})
}
