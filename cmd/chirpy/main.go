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
	"chirpy/internal/logging"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	cfg := config.Load()

	logger := logging.NewLogger(cfg)

	dbConn, err := sql.Open("postgres", cfg.Env.DBUrl)
	if err != nil {
		logger.Error("Error opening DB", "error", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	if err := dbConn.Ping(); err != nil {
		logger.Error("DB ping", "error", err)
		return
	}

	app := app.BuildApp(cfg, dbConn, logger)

	host := cfg.Env.Host
	port := cfg.Env.Port
	if port == "" {
		port = "8080"
	}
	addr := host + ":" + port

	srv := &http.Server{
		Addr:    addr,
		Handler: app,
	}

	var wg sync.WaitGroup

	serverErrCh := make(chan error)

	wg.Go(func() {
		logger.Info("Server starting", "addr", addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
	})

	sigCh := make(chan os.Signal, 1)

	signal.Notify(
		sigCh,
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer signal.Stop(sigCh)

	select {
	case err = <-serverErrCh:
		logger.Error("Server crashed", "error", err)
		return

	case sig := <-sigCh:
		logger.Info("Sys signal", "sig", sig)

		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err = srv.Shutdown(ctx); err != nil {
			logger.Error("Shutdown failed", "error", err)
			return
		}

		logger.Info("Server stopped")
	}

	wg.Wait()
}
