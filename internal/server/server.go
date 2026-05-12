package server

import (
	"chirpy/internal/handlers"
	"net/http"
)

type Server struct {
	mux           *http.ServeMux
	handlerAdmin  *handlers.HandlerAdmin
	handlerUsers  *handlers.HandlerUsers
	handlerChirps *handlers.HandlerChirps
	handlerApp    *handlers.HandlerApp
}

type Deps struct {
	HandlerAdmin  *handlers.HandlerAdmin
	HandlerUsers  *handlers.HandlerUsers
	HandlerChirps *handlers.HandlerChirps
	HandlerApp    *handlers.HandlerApp
}

func New(deps Deps) *http.ServeMux {
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
		handlerAdmin:  deps.HandlerAdmin,
		handlerUsers:  deps.HandlerUsers,
		handlerChirps: deps.HandlerChirps,
		handlerApp:    deps.HandlerApp,
	}

	s.registerRoutes()

	return s.mux
}
