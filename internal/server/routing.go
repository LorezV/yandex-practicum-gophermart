package server

import (
	"github.com/LorezV/go-diploma.git/internal/middlewares"
)

func (s *Server) registerRoutes() {
	authMiddleware := &middlewares.Authorization{Services: s.services}

	userGroup := s.core.Group("/api/user")
	userGroup.POST("/register", s.handler.Register)
	userGroup.POST("/login", s.handler.Login)

	authGroup := userGroup.Group("", authMiddleware.Handle)
	authGroup.POST("/orders", s.handler.PostOrders)
	authGroup.GET("/orders", s.handler.GetOrders)
	authGroup.GET("/balance", s.handler.GetBalance)
	authGroup.GET("/withdrawals", s.handler.GetWithdrawals)
	authGroup.POST("/balance/withdraw", s.handler.PostWithdraw)
}
