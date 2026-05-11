package main

import (
	"chirpy/internal/config"
	"chirpy/internal/database"
	"chirpy/internal/handlers"
	"chirpy/internal/server"
	"database/sql"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func newDB(dbUrl string) *database.Queries {
	dbConn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Panicf("Error opening DB: %s", err)
	}
	return database.New(dbConn)
}

func main() {
	godotenv.Load()

	cfg := config.Load()
	db := newDB(cfg.Env.DBUrl)

	handlerAdmin := handlers.NewHandlerAdmin(cfg, db)
	handlerUsers := handlers.NewHandlerUsers(cfg, db)
	handlerChirps := handlers.NewHandlerChirps(cfg, db)
	handlerApp := handlers.NewHandlerApp(cfg)

	server := server.New(server.Deps{
		HandlerAdmin:  handlerAdmin,
		HandlerUsers:  handlerUsers,
		HandlerChirps: handlerChirps,
		HandlerApp:    handlerApp,
	})

	log.Fatal(server.Run())
}
