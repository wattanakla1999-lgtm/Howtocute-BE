package dto

import (
	"nailly-back-end/internal/model"
	"time"
)

type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age" binding:"required,gt=0"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type UserResponse struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Phone      *string   `json:"phone,omitempty"`
	Age        int       `json:"age"`
	LineUserID *string   `json:"lineUserId,omitempty"`
	PictureURL string    `json:"pictureUrl,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CustomerMeResponse struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Phone      *string `json:"phone"`
	LineUserID *string `json:"lineUserId,omitempty"`
	PictureURL string  `json:"pictureUrl,omitempty"`
}

func ToUserResponse(user model.User) UserResponse {
	return UserResponse{
		ID:         user.ID,
		Name:       user.Name,
		Email:      user.Email,
		Phone:      user.Phone,
		Age:        user.Age,
		LineUserID: user.LineUserID,
		PictureURL: user.PictureURL,
		CreatedAt:  user.CreatedAt.In(thailandLocation),
		UpdatedAt:  user.UpdatedAt.In(thailandLocation),
	}
}

func ToUserResponses(users []model.User) []UserResponse {
	responses := make([]UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, ToUserResponse(user))
	}

	return responses
}

func (r CreateUserRequest) ToModel() model.User {
	return model.User{
		Name:  r.Name,
		Email: r.Email,
		Age:   r.Age,
	}
}

func (r UpdateUserRequest) ToModel() model.User {
	return model.User{
		Name:  r.Name,
		Email: r.Email,
		Age:   r.Age,
	}
}

func ToCustomerMeResponse(user model.User) CustomerMeResponse {
	return CustomerMeResponse{
		ID:         user.ID,
		Name:       user.Name,
		Phone:      user.Phone,
		LineUserID: user.LineUserID,
		PictureURL: user.PictureURL,
	}
}
