package service

import (
	"context"
	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/model"
	"nailly-back-end/internal/repository"
	"nailly-back-end/pkg/utils"
	"strings"
)

type ServicesService struct {
	repo          *repository.ServiceRepository
	imageUploader ImageUploader
}

func NewServicesService(repo *repository.ServiceRepository, uploaders ...ImageUploader) *ServicesService {
	var uploader ImageUploader
	if len(uploaders) > 0 {
		uploader = uploaders[0]
	}
	return &ServicesService{repo: repo, imageUploader: uploader}
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
	if err := s.storeServiceImage(&input); err != nil {
		return model.Service{}, err
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
	input.ID = service.ID
	if err := s.storeServiceImage(&input); err != nil {
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

func (s *ServicesService) storeServiceImage(input *model.Service) error {
	image := strings.TrimSpace(firstNonEmptyString(input.Img, input.ImageURL, input.ServiceImg))
	if image == "" {
		input.Img = ""
		input.ImageURL = ""
		input.ServiceImg = ""
		return nil
	}
	storedImage, err := s.storeCatalogImage("service-images", input.ID, image)
	if err != nil {
		return err
	}
	input.Img = storedImage
	input.ImageURL = storedImage
	input.ServiceImg = storedImage
	return nil
}

func (s *ServicesService) storeCatalogImage(folder string, entityID uint, image string) (string, error) {
	if !strings.HasPrefix(strings.TrimSpace(image), "data:") {
		return image, nil
	}
	if s.imageUploader == nil {
		return image, nil
	}
	uploadedURL, err := s.imageUploader.UploadImage(context.Background(), folder, entityID, image)
	if err != nil {
		return "", apperror.Internal("could not upload service image", err)
	}
	return uploadedURL, nil
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
