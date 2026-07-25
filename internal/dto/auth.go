package dto

import (
	"nailly-back-end/internal/model"
	"nailly-back-end/internal/service"
	"time"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type LoginResponse struct {
	Token     string        `json:"token"`
	TokenType string        `json:"tokenType"`
	ExpiresAt time.Time     `json:"expiresAt"`
	User      AdminResponse `json:"user"`
}

type LineLoginRequest struct {
	IDToken string `json:"idToken" binding:"required"`
}

type LineCustomerResponse struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	LineUserID *string `json:"lineUserId,omitempty"`
	PictureURL string  `json:"pictureUrl,omitempty"`
}

type LineLoginResponse struct {
	Token    string               `json:"token"`
	Customer LineCustomerResponse `json:"customer"`
}

func ToLoginResponse(result service.AuthResult) LoginResponse {
	return LoginResponse{
		Token: result.Token, TokenType: "Bearer", ExpiresAt: result.ExpiresAt,
		User: ToAdminResponse(result.Admin),
	}
}

func ToAdminResponse(admin model.Admin) AdminResponse {
	return AdminResponse{ID: admin.ID, Username: admin.Username, Name: admin.Name, Role: admin.Role}
}

func ToLineLoginResponse(result service.LineAuthResult) LineLoginResponse {
	return LineLoginResponse{
		Token:    result.Token,
		Customer: ToLineCustomerResponse(result.Customer),
	}
}

func ToLineCustomerResponse(customer model.User) LineCustomerResponse {
	return LineCustomerResponse{
		ID:         customer.ID,
		Name:       customer.Name,
		LineUserID: customer.LineUserID,
		PictureURL: customer.PictureURL,
	}
}
