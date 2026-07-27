package service

import (
	"errors"
	"strings"

	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/model"
	"nailly-back-end/internal/repository"

	"gorm.io/gorm"
)

type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) GetAllCategories() ([]model.Category, error) {
	return s.repo.GetAll()
}

func (s *CategoryService) CreateCategory(categoryID, name, slug string, displayOrder int) (model.Category, error) {
	categoryID = strings.TrimSpace(categoryID)
	name = strings.TrimSpace(name)
	if categoryID == "" || name == "" {
		return model.Category{}, apperror.BadRequest("categoryId and name are required", apperror.ErrValidation)
	}

	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	}

	category := model.Category{
		CategoryID:   categoryID,
		Name:         name,
		Slug:         slug,
		DisplayOrder: displayOrder,
	}

	if err := s.repo.Create(&category); err != nil {
		return model.Category{}, apperror.Internal("failed to create category", err)
	}

	return category, nil
}

func (s *CategoryService) UpdateCategory(id uint, name, slug string, displayOrder int) (model.Category, error) {
	category, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Category{}, apperror.NotFound("category not found", err)
		}
		return model.Category{}, apperror.Internal("failed to find category", err)
	}

	if name != "" {
		category.Name = strings.TrimSpace(name)
	}
	if slug != "" {
		category.Slug = strings.TrimSpace(slug)
	}
	category.DisplayOrder = displayOrder

	if err := s.repo.Update(&category); err != nil {
		return model.Category{}, apperror.Internal("failed to update category", err)
	}

	return category, nil
}

func (s *CategoryService) DeleteCategory(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return apperror.Internal("failed to delete category", err)
	}
	return nil
}
