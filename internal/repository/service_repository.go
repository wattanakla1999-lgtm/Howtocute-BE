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
	return r.db.Model(service).Updates(input).Error
}

func (r *ServiceRepository) Delete(service *model.Service) error {
	return r.db.Delete(service).Error
}
