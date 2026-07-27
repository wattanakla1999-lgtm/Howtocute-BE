package router

import (
	"nailly-back-end/internal/handler"
	"nailly-back-end/internal/repository"
	"nailly-back-end/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterServiceRoutes(api *gin.RouterGroup, db *gorm.DB, requireAdmin gin.HandlerFunc, imageUploader service.ImageUploader) {
	serviceRepository := repository.NewServiceRepository(db)
	serviceService := service.NewServicesService(serviceRepository, imageUploader)
	serviceHandler := handler.NewServicesHandler(serviceService)

	services := api.Group("/services")
	services.GET("", serviceHandler.GetServices)
	services.GET("/:id", serviceHandler.GetServiceByID)
	services.POST("", requireAdmin, serviceHandler.CreateService)
	services.PUT("/:id", requireAdmin, serviceHandler.UpdateService)
	services.DELETE("/:id", requireAdmin, serviceHandler.DeleteService)
}
