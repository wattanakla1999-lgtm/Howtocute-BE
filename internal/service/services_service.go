package service

import (
	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/model"
	"nailly-back-end/internal/repository"
	"nailly-back-end/pkg/utils"
)

type ServicesService struct {
	repo *repository.ServiceRepository
}

func NewServicesService(repo *repository.ServiceRepository) *ServicesService {
	return &ServicesService{repo: repo}
}

func (s *ServicesService) GetServices(filter repository.ServiceFilter, pagination utils.Pagination) ([]model.Service, int64, error) {
	return s.repo.FindAll(filter, pagination)
}

func (s *ServicesService) GetServiceByID(id string) (model.Service, error) {
	return s.repo.FindByID(id)
}

func (s *ServicesService) CreateService(input model.Service) (model.Service, error) {
	if input.ServiceName == "" {
		return model.Service{}, apperror.BadRequest("service name is required", apperror.ErrValidation)
	}
	if input.ServicePrice <= 0 {
		return model.Service{}, apperror.BadRequest("service price must be greater than 0", apperror.ErrValidation)
	}

	if err := s.repo.Create(&input); err != nil {
		return model.Service{}, err
	}

	return input, nil
}

func (s *ServicesService) UpdateService(id string, input model.Service) (model.Service, error) {
	service, err := s.repo.FindByID(id)
	if err != nil {
		return model.Service{}, err
	}

	if err := s.repo.Update(&service, input); err != nil {
		return model.Service{}, err
	}

	return service, nil
}

func (s *ServicesService) DeleteService(id string) error {
	service, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(&service)
}
