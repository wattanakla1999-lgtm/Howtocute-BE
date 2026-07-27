package router

import (
	"nailly-back-end/internal/handler"
	"nailly-back-end/internal/repository"
	"nailly-back-end/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterCategoryRoutes(api *gin.RouterGroup, db *gorm.DB, requireAdmin gin.HandlerFunc) {
	categoryRepository := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepository)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	categories := api.Group("/categories")
	categories.GET("", categoryHandler.GetCategories)
	categories.POST("", requireAdmin, categoryHandler.CreateCategory)
	categories.PUT("/:id", requireAdmin, categoryHandler.UpdateCategory)
	categories.DELETE("/:id", requireAdmin, categoryHandler.DeleteCategory)
}
