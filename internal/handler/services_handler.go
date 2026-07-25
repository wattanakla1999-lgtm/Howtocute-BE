package handler

import (
	"github.com/gin-gonic/gin"
	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/dto"
	"nailly-back-end/internal/repository"
	"nailly-back-end/internal/service"
	"nailly-back-end/pkg/utils"
	"net/http"
)

type ServicesHandler struct {
	service *service.ServicesService
}

func NewServicesHandler(service *service.ServicesService) *ServicesHandler {
	return &ServicesHandler{service: service}
}

func (h *ServicesHandler) GetServices(c *gin.Context) {
	filter := repository.ServiceFilter{
		ServiceName: c.Query("service_name"),
	}
	pagination := utils.NewPagination(c.DefaultQuery("page", "1"), c.DefaultQuery("limit", "10"))

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

// Removed GetUsersOlderThan method as it is not related to services

func (h *ServicesHandler) CreateService(c *gin.Context) {
	var input dto.CreateServiceRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, apperror.BadRequest("invalid request body", err))
		return
	}

	service, err := h.service.CreateService(input.ToModel())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToServiceResponse(service))
}

func (h *ServicesHandler) UpdateService(c *gin.Context) {
	var input dto.UpdateServiceRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, apperror.BadRequest("invalid request body", err))
		return
	}

	service, err := h.service.UpdateService(c.Param("id"), input.ToModel())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToServiceResponse(service))
}

func (h *ServicesHandler) DeleteService(c *gin.Context) {
	if err := h.service.DeleteService(c.Param("id")); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "service deleted"})
}
