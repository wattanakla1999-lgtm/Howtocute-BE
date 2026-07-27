package service

import (
	"context"
	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/model"
	"nailly-back-end/internal/repository"
	"nailly-back-end/pkg/utils"
	"strings"
)

type NailTechnicianService struct {
	repo          *repository.NailTechnicianRepository
	imageUploader ImageUploader
}

func NewNailTechnicianService(repo *repository.NailTechnicianRepository, uploaders ...ImageUploader) *NailTechnicianService {
	var uploader ImageUploader
	if len(uploaders) > 0 {
		uploader = uploaders[0]
	}
	return &NailTechnicianService{repo: repo, imageUploader: uploader}
}

func (s *NailTechnicianService) GetNailTechnicians(filter repository.NailTechnicianFilter, pagination utils.Pagination) ([]model.NailTechnician, int64, error) {
	return s.repo.FindAll(filter, pagination)
}

func (s *NailTechnicianService) GetNailTechnicianByID(id string) (model.NailTechnician, error) {
	return s.repo.FindByID(id)
}

func (s *NailTechnicianService) CreateNailTechnician(input model.NailTechnician) (model.NailTechnician, error) {
	if input.TechnicianName == "" {
		return model.NailTechnician{}, apperror.BadRequest("technician name is required", apperror.ErrValidation)
	}
	if input.ExperienceYears < 0 {
		return model.NailTechnician{}, apperror.BadRequest("experience years must be greater than or equal to 0", apperror.ErrValidation)
	}
	if err := s.storeTechnicianImage(&input); err != nil {
		return model.NailTechnician{}, err
	}

	if err := s.repo.Create(&input); err != nil {
		return model.NailTechnician{}, err
	}

	return input, nil
}

func (s *NailTechnicianService) UpdateNailTechnician(id string, input model.NailTechnician) (model.NailTechnician, error) {
	technician, err := s.repo.FindByID(id)
	if err != nil {
		return model.NailTechnician{}, err
	}
	input.ID = technician.ID
	if err := s.storeTechnicianImage(&input); err != nil {
		return model.NailTechnician{}, err
	}

	if err := s.repo.Update(&technician, input); err != nil {
		return model.NailTechnician{}, err
	}

	return technician, nil
}

func (s *NailTechnicianService) DeleteNailTechnician(id string) error {
	technician, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(&technician)
}

func (s *NailTechnicianService) storeTechnicianImage(input *model.NailTechnician) error {
	image := strings.TrimSpace(firstNonEmptyString(input.ProfileImg, input.AvatarURL))
	if image == "" {
		input.ProfileImg = ""
		input.AvatarURL = ""
		return nil
	}
	storedImage, err := s.storeCatalogImage("technician-profiles", input.ID, image)
	if err != nil {
		return err
	}
	input.ProfileImg = storedImage
	input.AvatarURL = storedImage
	return nil
}

func (s *NailTechnicianService) storeCatalogImage(folder string, entityID uint, image string) (string, error) {
	if !strings.HasPrefix(strings.TrimSpace(image), "data:") {
		return image, nil
	}
	if s.imageUploader == nil {
		return image, nil
	}
	uploadedURL, err := s.imageUploader.UploadImage(context.Background(), folder, entityID, image)
	if err != nil {
		return "", apperror.Internal("could not upload technician profile image", err)
	}
	return uploadedURL, nil
}
