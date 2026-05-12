package server

import (
	"chirpy/internal/config"
	"chirpy/internal/handlers"
	"log/slog"
	"net/http"
)

type Server struct {
	mux           *http.ServeMux
	cfg           *config.Config
	logger        *slog.Logger
	handlerAdmin  *handlers.HandlerAdmin
	handlerUsers  *handlers.HandlerUsers
	handlerChirps *handlers.HandlerChirps
	handlerApp    *handlers.HandlerApp
}

type Deps struct {
	Config        *config.Config
	Logger        *slog.Logger
	HandlerAdmin  *handlers.HandlerAdmin
	HandlerUsers  *handlers.HandlerUsers
	HandlerChirps *handlers.HandlerChirps
	HandlerApp    *handlers.HandlerApp
}

func New(deps Deps) http.Handler {
	if deps.Config == nil {
		panic("config is nil")
	}
	if deps.Logger == nil {
		panic("logger is nil")
	}
	if deps.HandlerAdmin == nil {
		panic("admin handler is nil")
	}
	if deps.HandlerUsers == nil {
		panic("users handler is nil")
	}
	if deps.HandlerChirps == nil {
		panic("chirps handler is nil")
	}
	if deps.HandlerApp == nil {
		panic("app handler is nil")
	}

	s := &Server{
		mux:           http.NewServeMux(),
		cfg:           deps.Config,
		logger:        deps.Logger,
		handlerAdmin:  deps.HandlerAdmin,
		handlerUsers:  deps.HandlerUsers,
		handlerChirps: deps.HandlerChirps,
		handlerApp:    deps.HandlerApp,
	}

	s.registerRoutes()

	return s.getHandler()
}
