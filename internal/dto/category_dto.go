package dto

import (
	"nailly-back-end/internal/model"
	"time"
)

type CreateCategoryRequest struct {
	CategoryID   string `json:"categoryId" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Slug         string `json:"slug"`
	DisplayOrder int    `json:"displayOrder"`
}

type UpdateCategoryRequest struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	DisplayOrder int    `json:"displayOrder"`
}

type CategoryResponse struct {
	ID           uint      `json:"id"`
	CategoryID   string    `json:"categoryId"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	DisplayOrder int       `json:"displayOrder"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func ToCategoryResponse(category model.Category) CategoryResponse {
	return CategoryResponse{
		ID:           category.ID,
		CategoryID:   category.CategoryID,
		Name:         category.Name,
		Slug:         category.Slug,
		DisplayOrder: category.DisplayOrder,
		CreatedAt:    category.CreatedAt.In(thailandLocation),
		UpdatedAt:    category.UpdatedAt.In(thailandLocation),
	}
}

func ToCategoryResponses(categories []model.Category) []CategoryResponse {
	responses := make([]CategoryResponse, 0, len(categories))
	for _, category := range categories {
		responses = append(responses, ToCategoryResponse(category))
	}
	return responses
}
