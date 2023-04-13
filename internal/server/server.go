package server

import (
	"context"
	"github.com/LorezV/go-diploma.git/internal/handlers"
	"github.com/LorezV/go-diploma.git/internal/services"
	"github.com/labstack/echo/v4"
)

type Server struct {
	core     *echo.Echo
	handler  *handlers.Handler
	services *services.Services
	address  string
}

func NewServer(address string, services *services.Services) *Server {
	s := &Server{
		core:     echo.New(),
		handler:  handlers.MakeHandler(services),
		services: services,
		address:  address,
	}
	s.registerRoutes()

	return s
}

func (s *Server) Run() error {
	return s.core.Start(s.address)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.core.Shutdown(ctx)
}
