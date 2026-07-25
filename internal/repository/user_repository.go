package repository

import (
	"nailly-back-end/internal/model"
	"nailly-back-end/pkg/utils"

	"gorm.io/gorm"
)

type UserFilter struct {
	Name  string
	Email string
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindAll(filter UserFilter, pagination utils.Pagination) ([]model.User, int64, error) {
	var users []model.User

	query := r.db.Model(&model.User{})
	query = utils.ApplyLikeFilters(query, map[string]string{
		"name":  filter.Name,
		"email": filter.Email,
	})

	total, err := utils.Paginate(query, pagination, &users)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) FindByID(id string) (model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return user, err
}

func (r *UserRepository) FindByEmail(email string) (model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return user, err
}

func (r *UserRepository) FindByLineUserID(lineUserID string) (model.User, error) {
	var user model.User
	err := r.db.Where("line_user_id = ?", lineUserID).First(&user).Error
	return user, err
}

func (r *UserRepository) FindOlderThan(age int) ([]model.User, error) {
	var users []model.User
	err := r.db.Where("age > ?", age).Find(&users).Error
	return users, err
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Update(user *model.User, input model.User) error {
	return r.db.Model(user).Updates(input).Error
}

func (r *UserRepository) Save(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(user *model.User) error {
	return r.db.Delete(user).Error
}
