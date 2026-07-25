package handler

import (
	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/dto"
	"nailly-back-end/internal/middleware"
	"nailly-back-end/internal/model"
	"nailly-back-end/internal/repository"
	"nailly-back-end/internal/service"
	"nailly-back-end/pkg/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	service *service.BookingService
}

func NewBookingHandler(bookingService *service.BookingService) *BookingHandler {
	return &BookingHandler{service: bookingService}
}

func (h *BookingHandler) GetBookings(c *gin.Context) {
	filter, err := bookingFilterFromQuery(c)
	if err != nil {
		respondError(c, err)
		return
	}
	pagination := utils.NewPagination(c.DefaultQuery("page", "1"), c.DefaultQuery("limit", "6"))

	bookings, total, err := h.service.GetBookings(filter, pagination)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data: dto.ToBookingResponses(bookings), Page: pagination.Page,
		Limit: pagination.Limit, Total: total,
	})
}

func (h *BookingHandler) GetCustomerBookings(c *gin.Context) {
	phone := strings.TrimSpace(c.Query("phone"))
	if phone == "" {
		respondError(c, apperror.BadRequest("phone is required", apperror.ErrValidation))
		return
	}
	pagination := utils.NewPagination(c.DefaultQuery("page", "1"), c.DefaultQuery("limit", "100"))
	bookings, total, err := h.service.GetBookings(repository.BookingFilter{CustomerPhone: phone}, pagination)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data: dto.ToBookingResponses(bookings), Page: pagination.Page,
		Limit: pagination.Limit, Total: total,
	})
}

func (h *BookingHandler) GetMyBookings(c *gin.Context) {
	customerID, err := customerIDFromContext(c)
	if err != nil {
		respondError(c, err)
		return
	}

	pagination := utils.NewPagination(c.DefaultQuery("page", "1"), c.DefaultQuery("limit", "100"))
	bookings, total, err := h.service.GetBookings(repository.BookingFilter{UserID: &customerID}, pagination)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data: dto.ToBookingResponses(bookings), Page: pagination.Page,
		Limit: pagination.Limit, Total: total,
	})
}

func (h *BookingHandler) GetBookingByID(c *gin.Context) {
	id, err := bookingIDFromParam(c)
	if err != nil {
		respondError(c, err)
		return
	}
	booking, err := h.service.GetBookingByID(id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToBookingResponse(booking))
}

func (h *BookingHandler) GetBusySlots(c *gin.Context) {
	rawDate := c.Query("date")
	date, err := time.ParseInLocation("2006-01-02", rawDate, time.Local)
	if err != nil {
		respondError(c, apperror.BadRequest("date must use YYYY-MM-DD format", err))
		return
	}
	technicianID, err := optionalUintQuery(c, "technicianId")
	if err != nil {
		respondError(c, err)
		return
	}
	serviceID, err := optionalUintQuery(c, "serviceId")
	if err != nil {
		respondError(c, err)
		return
	}
	busySlots, err := h.service.GetBusySlots(date, technicianID, serviceID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"busySlots": busySlots})
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var request dto.CreateBookingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperror.BadRequest("invalid request body", err))
		return
	}
	booking, err := h.service.CreateBooking(service.CreateBookingInput{
		UserID: request.UserID, ServiceID: request.ServiceID, TechnicianID: request.TechnicianID,
		StartAt: request.StartAt, EndAt: request.EndAt, CustomerName: request.CustomerName,
		CustomerPhone: request.CustomerPhone, PaymentMethod: request.PaymentMethod, Note: request.Note,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToBookingResponse(booking))
}

func (h *BookingHandler) CreateCustomerBooking(c *gin.Context) {
	customerID, err := customerIDFromContext(c)
	if err != nil {
		respondError(c, err)
		return
	}

	var request dto.CreateBookingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperror.BadRequest("invalid request body", err))
		return
	}

	booking, err := h.service.CreateBooking(service.CreateBookingInput{
		UserID: &customerID, ServiceID: request.ServiceID, TechnicianID: request.TechnicianID,
		StartAt: request.StartAt, EndAt: request.EndAt, CustomerName: request.CustomerName,
		CustomerPhone: request.CustomerPhone, PaymentMethod: request.PaymentMethod, Note: request.Note,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToBookingResponse(booking))
}

func (h *BookingHandler) UpdateBooking(c *gin.Context) {
	id, err := bookingIDFromParam(c)
	if err != nil {
		respondError(c, err)
		return
	}
	var request dto.UpdateBookingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperror.BadRequest("invalid request body", err))
		return
	}
	booking, err := h.service.UpdateBooking(id, service.UpdateBookingInput{
		UserID: request.UserID.Value, UserIDSet: request.UserID.Set, ServiceID: request.ServiceID,
		TechnicianID: request.TechnicianID.Value, TechnicianIDSet: request.TechnicianID.Set,
		StartAt: request.StartAt, EndAt: request.EndAt, CustomerName: request.CustomerName,
		CustomerPhone: request.CustomerPhone, PaymentMethod: request.PaymentMethod, Note: request.Note,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToBookingResponse(booking))
}

func (h *BookingHandler) UpdateBookingStatus(c *gin.Context) {
	id, err := bookingIDFromParam(c)
	if err != nil {
		respondError(c, err)
		return
	}
	var request dto.UpdateBookingStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperror.BadRequest("invalid request body", err))
		return
	}
	booking, err := h.service.UpdateBookingStatus(id, request.Status, request.CancelReason)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToBookingResponse(booking))
}

func (h *BookingHandler) CancelCustomerBooking(c *gin.Context) {
	customerID, err := customerIDFromContext(c)
	if err != nil {
		respondError(c, err)
		return
	}
	id, err := bookingIDFromParam(c)
	if err != nil {
		respondError(c, err)
		return
	}
	var request dto.CancelBookingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperror.BadRequest("invalid request body", err))
		return
	}
	booking, err := h.service.CancelCustomerBooking(id, customerID, request.CancelReason)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToBookingResponse(booking))
}

func (h *BookingHandler) DeleteBooking(c *gin.Context) {
	id, err := bookingIDFromParam(c)
	if err != nil {
		respondError(c, err)
		return
	}
	if err := h.service.DeleteBooking(id); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "booking deleted"})
}

func bookingIDFromParam(c *gin.Context) (uint, error) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		return 0, apperror.BadRequest("invalid booking id", apperror.ErrValidation)
	}
	return uint(value), nil
}

func customerIDFromContext(c *gin.Context) (uint, error) {
	claims, ok := middleware.CustomerClaimsFromContext(c)
	if !ok {
		return 0, apperror.Unauthorized("invalid customer token", apperror.ErrValidation)
	}
	value, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil || value == 0 {
		return 0, apperror.Unauthorized("invalid customer token", err)
	}
	return uint(value), nil
}

func bookingFilterFromQuery(c *gin.Context) (repository.BookingFilter, error) {
	filter := repository.BookingFilter{Status: model.BookingStatus(c.Query("status"))}
	var err error
	if filter.UserID, err = optionalUintQuery(c, "userId"); err != nil {
		return filter, err
	}
	if filter.ServiceID, err = optionalUintQuery(c, "serviceId"); err != nil {
		return filter, err
	}
	if filter.TechnicianID, err = optionalUintQuery(c, "technicianId"); err != nil {
		return filter, err
	}
	if filter.DateFrom, err = optionalDateQuery(c, "dateFrom", false); err != nil {
		return filter, err
	}
	if filter.DateTo, err = optionalDateQuery(c, "dateTo", true); err != nil {
		return filter, err
	}
	return filter, nil
}

func optionalUintQuery(c *gin.Context, name string) (*uint, error) {
	raw := c.Query(name)
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		return nil, apperror.BadRequest(name+" must be a positive integer", apperror.ErrValidation)
	}
	parsed := uint(value)
	return &parsed, nil
}

func optionalDateQuery(c *gin.Context, name string, endOfDay bool) (*time.Time, error) {
	raw := c.Query(name)
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err == nil {
		return &parsed, nil
	}
	parsed, err = time.ParseInLocation("2006-01-02", raw, time.Local)
	if err != nil {
		return nil, apperror.BadRequest(name+" must use YYYY-MM-DD or RFC3339 format", err)
	}
	if endOfDay {
		parsed = parsed.Add(24*time.Hour - time.Nanosecond)
	}
	return &parsed, nil
}
