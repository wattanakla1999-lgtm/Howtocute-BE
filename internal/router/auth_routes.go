package router

import (
	"nailly-back-end/internal/handler"
	"nailly-back-end/internal/repository"
	"nailly-back-end/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterAuthRoutes(api *gin.RouterGroup, db *gorm.DB, jwtManager *service.JWTManager, customerJWTManager *service.CustomerJWTManager, lineLoginChannelID string, requireAdmin gin.HandlerFunc) {
	authRepository := repository.NewAuthRepository(db)
	authService := service.NewAuthService(authRepository, jwtManager)
	userRepository := repository.NewUserRepository(db)
	lineAuthService := service.NewLineAuthService(userRepository, service.NewHTTPLineTokenVerifier(nil), customerJWTManager, lineLoginChannelID)
	authHandler := handler.NewAuthHandler(authService, lineAuthService)

	auth := api.Group("/auth")
	auth.POST("/login", authHandler.Login)
	auth.POST("/line", authHandler.LineLogin)
	auth.GET("/me", requireAdmin, authHandler.Me)
}
