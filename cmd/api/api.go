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

func main() {
	godotenv.Load()

	cfg := config.Load()

	dbConn, err := sql.Open("postgres", cfg.Env.DBUrl)
	if err != nil {
		log.Fatalf("Error opening DB: %s", err)
		return
	}
	db := database.New(dbConn)

	handlerAdmin := handlers.NewHandlerAdmin(cfg, db)
	handlerUsers := handlers.NewHandlerUsers(db)
	handlerChirps := handlers.NewHandlerChirps(db)
	handlerApp := handlers.NewHandlerApp(cfg)

	server := server.New(server.Deps{
		HandlerAdmin:  handlerAdmin,
		HandlerUsers:  handlerUsers,
		HandlerChirps: handlerChirps,
		HandlerApp:    handlerApp,
	})

	log.Fatal(server.Run())
}
