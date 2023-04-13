package server

import (
	"github.com/LorezV/go-diploma.git/internal/middlewares"
)

func (s *Server) registerRoutes() {
	apiGroup := s.core.Group("/api", middlewares.JSONGuard)

	userGroup := apiGroup.Group("/user")
	userGroup.POST("/register", s.handler.Register)
	userGroup.POST("/login", s.handler.Login)

	authMiddleware := &middlewares.Authorization{Services: s.services}
	authGroup := userGroup.Group("", authMiddleware.Handle)
	authGroup.POST("/balance/withdraw", s.handler.PostWithdraw)

	simpleAuthGroup := s.core.Group("/api/user", authMiddleware.Handle)
	simpleAuthGroup.POST("/orders", s.handler.PostOrders)
	simpleAuthGroup.GET("/orders", s.handler.GetOrders)
	simpleAuthGroup.GET("/balance", s.handler.GetBalance)
	simpleAuthGroup.GET("/withdrawals", s.handler.GetWithdrawals)
}
