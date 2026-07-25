package router

import (
	"nailly-back-end/internal/handler"
	"nailly-back-end/internal/repository"
	"nailly-back-end/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterCustomerRoutes(api *gin.RouterGroup, db *gorm.DB, requireCustomer gin.HandlerFunc) {
	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	customers := api.Group("/customers")
	customers.GET("/me", requireCustomer, userHandler.GetCustomerMe)
}
