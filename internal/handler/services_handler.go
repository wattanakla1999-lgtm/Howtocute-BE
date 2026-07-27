package handler

import (
	"net/http"
	"strconv"

	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/dto"
	"nailly-back-end/internal/repository"
	"nailly-back-end/internal/service"
	"nailly-back-end/pkg/utils"

	"github.com/gin-gonic/gin"
)

type ServicesHandler struct {
	service *service.ServicesService
}

func NewServicesHandler(service *service.ServicesService) *ServicesHandler {
	return &ServicesHandler{service: service}
}

func (h *ServicesHandler) GetServices(c *gin.Context) {
	var categoryIDPtr *uint
	if catIDStr := c.Query("category_id"); catIDStr != "" {
		if val, err := strconv.ParseUint(catIDStr, 10, 32); err == nil {
			uVal := uint(val)
			categoryIDPtr = &uVal
		}
	}

	filter := repository.ServiceFilter{
		ServiceName: c.Query("service_name"),
		Category:    c.Query("category"),
		CategoryID:  categoryIDPtr,
	}

	// If page/limit parameters are omitted, set limit to a large number to return all services at once
	limit := c.Query("limit")
	if limit == "" {
		limit = "1000"
	}

	pagination := utils.NewPagination(c.DefaultQuery("page", "1"), limit)

	services, total, err := h.service.GetServices(filter, pagination)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:  dto.ToServiceResponses(services),
		Page:  pagination.Page,
		Limit: pagination.Limit,
		Total: total,
	})
}

func (h *ServicesHandler) GetServiceByID(c *gin.Context) {
	service, err := h.service.GetServiceByID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToServiceResponse(service))
}

func (h *ServicesHandler) CreateService(c *gin.Context) {
	var request dto.CreateServiceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperror.BadRequest("invalid request body", err))
		return
	}

	serviceModel := request.ToModel()
	service, err := h.service.CreateService(serviceModel)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToServiceResponse(service))
}

func (h *ServicesHandler) UpdateService(c *gin.Context) {
	var request dto.UpdateServiceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperror.BadRequest("invalid request body", err))
		return
	}

	service, err := h.service.UpdateService(c.Param("id"), request.ToModel())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToServiceResponse(service))
}

func (h *ServicesHandler) DeleteService(c *gin.Context) {
	err := h.service.DeleteService(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service deleted successfully"})
}
