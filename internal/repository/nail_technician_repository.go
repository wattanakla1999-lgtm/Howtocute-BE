package repository

import (
	"nailly-back-end/internal/model"
	"nailly-back-end/pkg/utils"

	"gorm.io/gorm"
)

type NailTechnicianFilter struct {
	TechnicianName string
	Phone          string
	Specialty      string
}

type NailTechnicianRepository struct {
	db *gorm.DB
}

func NewNailTechnicianRepository(db *gorm.DB) *NailTechnicianRepository {
	return &NailTechnicianRepository{db: db}
}

func (r *NailTechnicianRepository) FindAll(filter NailTechnicianFilter, pagination utils.Pagination) ([]model.NailTechnician, int64, error) {
	var technicians []model.NailTechnician

	query := r.db.Model(&model.NailTechnician{})
	query = utils.ApplyLikeFilters(query, map[string]string{
		"technician_name": filter.TechnicianName,
		"phone":           filter.Phone,
		"specialty":       filter.Specialty,
	})

	total, err := utils.Paginate(query, pagination, &technicians)
	if err != nil {
		return nil, 0, err
	}

	return technicians, total, nil
}

func (r *NailTechnicianRepository) FindByID(id string) (model.NailTechnician, error) {
	var technician model.NailTechnician
	err := r.db.First(&technician, id).Error
	return technician, err
}

func (r *NailTechnicianRepository) Create(technician *model.NailTechnician) error {
	return r.db.Create(technician).Error
}

func (r *NailTechnicianRepository) Update(technician *model.NailTechnician, input model.NailTechnician) error {
	if err := r.db.Model(technician).Updates(map[string]any{
		"technician_name":  input.TechnicianName,
		"phone":            input.Phone,
		"experience_years": input.ExperienceYears,
		"specialty":        input.Specialty,
		"profile_img":      input.ProfileImg,
		"avatar_url":       input.AvatarURL,
		"active":           input.Active,
		"bio":              input.Bio,
	}).Error; err != nil {
		return err
	}
	technician.TechnicianName = input.TechnicianName
	technician.Phone = input.Phone
	technician.ExperienceYears = input.ExperienceYears
	technician.Specialty = input.Specialty
	technician.ProfileImg = input.ProfileImg
	technician.AvatarURL = input.AvatarURL
	technician.Active = input.Active
	technician.Bio = input.Bio
	return nil
}

func (r *NailTechnicianRepository) Delete(technician *model.NailTechnician) error {
	return r.db.Delete(technician).Error
}
