package main

import (
	"chirpy/internal/app"
	"chirpy/internal/config"
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

func main() {
	godotenv.Load()
	cfg := config.Load()
	dbConn := newDB(cfg.Env.DBUrl)

	app := app.BuildApp(cfg, dbConn)

	log.Fatal(http.ListenAndServe(":8080", app))
}
