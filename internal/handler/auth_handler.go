package handler

import (
	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/dto"
	"nailly-back-end/internal/middleware"
	"nailly-back-end/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service     *service.AuthService
	lineService *service.LineAuthService
}

func NewAuthHandler(authService *service.AuthService, lineService *service.LineAuthService) *AuthHandler {
	return &AuthHandler{service: authService, lineService: lineService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request dto.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperror.BadRequest("username and password are required", err))
		return
	}
	result, err := h.service.Login(request.Username, request.Password)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToLoginResponse(result))
}

func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := middleware.AdminClaimsFromContext(c)
	if !ok {
		respondError(c, apperror.Unauthorized("กรุณาเข้าสู่ระบบ", apperror.ErrValidation))
		return
	}
	c.JSON(http.StatusOK, dto.AdminResponse{
		ID: claims.AdminID, Username: claims.Username, Name: claims.Name, Role: claims.Role,
	})
}

func (h *AuthHandler) LineLogin(c *gin.Context) {
	var request dto.LineLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperror.BadRequest("idToken is required", err))
		return
	}
	result, err := h.lineService.Login(c.Request.Context(), request.IDToken)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToLineLoginResponse(result))
}
