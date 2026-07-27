package repository

import (
	"nailly-back-end/internal/model"
	"nailly-back-end/pkg/utils"

	"gorm.io/gorm"
)

type ServiceFilter struct {
	ServiceName string
}

type ServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) FindAll(filter ServiceFilter, pagination utils.Pagination) ([]model.Service, int64, error) {
	var services []model.Service

	query := r.db.Model(&model.Service{})
	query = utils.ApplyLikeFilters(query, map[string]string{
		"service_name": filter.ServiceName,
	})

	total, err := utils.Paginate(query, pagination, &services)
	if err != nil {
		return nil, 0, err
	}

	return services, total, nil
}

func (r *ServiceRepository) FindByID(id string) (model.Service, error) {
	var service model.Service
	err := r.db.First(&service, id).Error
	return service, err
}

func (r *ServiceRepository) Create(service *model.Service) error {
	return r.db.Create(service).Error
}

func (r *ServiceRepository) Update(service *model.Service, input model.Service) error {
	if err := r.db.Model(service).Updates(map[string]any{
		"service_name":  input.ServiceName,
		"service_price": input.ServicePrice,
		"duration":      input.Duration,
		"image_url":     input.ImageURL,
		"img":           input.Img,
		"service_img":   input.ServiceImg,
		"popular":       input.Popular,
		"description":   input.Description,
	}).Error; err != nil {
		return err
	}
	service.ServiceName = input.ServiceName
	service.ServicePrice = input.ServicePrice
	service.Duration = input.Duration
	service.ImageURL = input.ImageURL
	service.Img = input.Img
	service.ServiceImg = input.ServiceImg
	service.Popular = input.Popular
	service.Description = input.Description
	return nil
}

func (r *ServiceRepository) Delete(service *model.Service) error {
	return r.db.Delete(service).Error
}
