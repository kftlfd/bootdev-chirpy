//	@title			Chirpy API
//	@version		1.0
//	@description	Chirpy backend API

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				JWT access token. Example: "Bearer {JWT}"

//	@securityDefinitions.apikey	RefreshAuth
//	@in							header
//	@name						Authorization
//	@description				Refresh token. Example: "Bearer {refresh_token}"

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description				API key. Example: "ApiKey {api_key}"

package main

import (
	_ "chirpy/docs"
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

	host := cfg.Env.Host
	port := cfg.Env.Port
	if port == "" {
		port = "8080"
	}
	addr := host + ":" + port
	log.Print("Listening on " + addr)

	log.Fatal(http.ListenAndServe(addr, app))
}
