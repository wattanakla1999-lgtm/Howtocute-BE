package router

import (
	"nailly-back-end/internal/handler"
	"nailly-back-end/internal/repository"
	"nailly-back-end/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterNailTechnicianRoutes(api *gin.RouterGroup, db *gorm.DB, requireAdmin gin.HandlerFunc, imageUploader service.ImageUploader) {
	nailTechnicianRepository := repository.NewNailTechnicianRepository(db)
	nailTechnicianService := service.NewNailTechnicianService(nailTechnicianRepository, imageUploader)
	nailTechnicianHandler := handler.NewNailTechnicianHandler(nailTechnicianService)

	nailTechnicians := api.Group("/nail_technician")
	nailTechnicians.GET("", nailTechnicianHandler.GetNailTechnicians)
	nailTechnicians.GET("/:id", nailTechnicianHandler.GetNailTechnicianByID)
	nailTechnicians.POST("", requireAdmin, nailTechnicianHandler.CreateNailTechnician)
	nailTechnicians.PUT("/:id", requireAdmin, nailTechnicianHandler.UpdateNailTechnician)
	nailTechnicians.DELETE("/:id", requireAdmin, nailTechnicianHandler.DeleteNailTechnician)
}
