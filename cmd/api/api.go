package main

import (
	"chirpy/internal/config"
	"chirpy/internal/database"
	"chirpy/internal/handlers"
	"chirpy/internal/metrics"
	"chirpy/internal/server"
	"database/sql"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func newDB(dbUrl string) *sql.DB {
	dbConn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Panicf("Error opening DB: %s", err)
	}
	return dbConn
}

func buildApp(
	cfg *config.Config,
	dbConn *sql.DB,
) *http.ServeMux {
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

func main() {
	godotenv.Load()
	cfg := config.Load()
	dbConn := newDB(cfg.Env.DBUrl)

	app := buildApp(cfg, dbConn)

	log.Fatal(http.ListenAndServe(":8080", app))
}
