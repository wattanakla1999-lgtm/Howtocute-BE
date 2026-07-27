package handler

import (
	"net/http"
	"strconv"

	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/dto"
	"nailly-back-end/internal/service"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	service *service.CategoryService
}

func NewCategoryHandler(service *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) GetCategories(c *gin.Context) {
	categories, err := h.service.GetAllCategories()
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToCategoryResponses(categories))
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var request dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperror.BadRequest("invalid request body", err))
		return
	}

	category, err := h.service.CreateCategory(request.CategoryID, request.Name, request.Slug, request.DisplayOrder)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToCategoryResponse(category))
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondError(c, apperror.BadRequest("invalid category id", err))
		return
	}

	var request dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperror.BadRequest("invalid request body", err))
		return
	}

	category, err := h.service.UpdateCategory(uint(id), request.Name, request.Slug, request.DisplayOrder)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCategoryResponse(category))
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondError(c, apperror.BadRequest("invalid category id", err))
		return
	}

	if err := h.service.DeleteCategory(uint(id)); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
