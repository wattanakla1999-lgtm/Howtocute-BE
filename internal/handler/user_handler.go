package handler

import (
	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/dto"
	"nailly-back-end/internal/middleware"
	"nailly-back-end/internal/model"
	"nailly-back-end/internal/repository"
	"nailly-back-end/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service UserService
}

type UserService interface {
	GetUsers(filter repository.UserFilter, pagination utils.Pagination) ([]model.User, int64, error)
	GetUserByID(id string) (model.User, error)
	GetUserByEmail(email string) (model.User, error)
	GetUsersOlderThan(age int) ([]model.User, error)
	CreateUser(input model.User) (model.User, error)
	UpdateUser(id string, input model.User) (model.User, error)
	DeleteUser(id string) error
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	filter := repository.UserFilter{
		Name:  c.Query("name"),
		Email: c.Query("email"),
	}
	pagination := utils.NewPagination(c.DefaultQuery("page", "1"), c.DefaultQuery("limit", "10"))

	users, total, err := h.service.GetUsers(filter, pagination)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:  dto.ToUserResponses(users),
		Page:  pagination.Page,
		Limit: pagination.Limit,
		Total: total,
	})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	user, err := h.service.GetUserByID(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	user, err := h.service.GetUserByEmail(c.Param("email"))
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

func (h *UserHandler) GetUsersOlderThan(c *gin.Context) {
	age, err := strconv.Atoi(c.Param("age"))
	if err != nil {
		respondError(c, apperror.BadRequest("invalid age", err))
		return
	}

	users, err := h.service.GetUsersOlderThan(age)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToUserResponses(users))
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var input dto.CreateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, apperror.BadRequest("invalid request body", err))
		return
	}

	user, err := h.service.CreateUser(input.ToModel())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToUserResponse(user))
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var input dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, apperror.BadRequest("invalid request body", err))
		return
	}

	user, err := h.service.UpdateUser(c.Param("id"), input.ToModel())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	if err := h.service.DeleteUser(c.Param("id")); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func (h *UserHandler) GetCustomerMe(c *gin.Context) {
	claims, ok := middleware.CustomerClaimsFromContext(c)
	if !ok {
		respondError(c, apperror.Unauthorized("กรุณาเข้าสู่ระบบ", apperror.ErrValidation))
		return
	}

	user, err := h.service.GetUserByID(claims.Subject)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToCustomerMeResponse(user))
}
