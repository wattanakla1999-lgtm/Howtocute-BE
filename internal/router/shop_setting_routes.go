package router

import (
	"nailly-back-end/internal/handler"
	"nailly-back-end/internal/repository"
	"nailly-back-end/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterShopSettingRoutes(api *gin.RouterGroup, db *gorm.DB, requireAdmin gin.HandlerFunc, qrCodeUploader service.ImageUploader) {
	settingRepository := repository.NewShopSettingRepository(db)
	settingService := service.NewShopSettingService(settingRepository, qrCodeUploader)
	settingHandler := handler.NewShopSettingHandler(settingService)

	settings := api.Group("/settings")
	settings.GET("", settingHandler.GetSettings)
	settings.Use(requireAdmin)
	settings.PUT("", settingHandler.UpdateSettings)
}
