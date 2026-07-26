package service

import (
	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/model"
	"strings"
	"time"
)

type ShopSettingStore interface {
	Get() (model.ShopSetting, error)
	Save(setting *model.ShopSetting) error
}

type UpdateShopSettingInput struct {
	ShopStatus      string
	OpenTime        string
	CloseTime       string
	ShopPhone       string
	PromptPayNumber string
	AccountName     string
	BankName        string
	DepositAmount   *float64
}

type ShopSettingService struct {
	repo ShopSettingStore
}

func NewShopSettingService(repo ShopSettingStore) *ShopSettingService {
	return &ShopSettingService{repo: repo}
}

func (s *ShopSettingService) GetSettings() (model.ShopSetting, error) {
	return s.repo.Get()
}

func (s *ShopSettingService) UpdateSettings(input UpdateShopSettingInput) (model.ShopSetting, error) {
	input.ShopStatus = strings.TrimSpace(input.ShopStatus)
	input.OpenTime = strings.TrimSpace(input.OpenTime)
	input.CloseTime = strings.TrimSpace(input.CloseTime)
	input.ShopPhone = strings.TrimSpace(input.ShopPhone)
	input.PromptPayNumber = strings.TrimSpace(input.PromptPayNumber)
	input.AccountName = strings.TrimSpace(input.AccountName)
	input.BankName = strings.TrimSpace(input.BankName)

	if input.ShopStatus != "open" && input.ShopStatus != "closed" {
		return model.ShopSetting{}, apperror.BadRequest("shopStatus must be open or closed", apperror.ErrValidation)
	}
	openTime, err := time.Parse("15:04", input.OpenTime)
	if err != nil {
		return model.ShopSetting{}, apperror.BadRequest("openTime must use HH:MM format", err)
	}
	closeTime, err := time.Parse("15:04", input.CloseTime)
	if err != nil {
		return model.ShopSetting{}, apperror.BadRequest("closeTime must use HH:MM format", err)
	}
	if !openTime.Before(closeTime) {
		return model.ShopSetting{}, apperror.BadRequest("openTime must be before closeTime", apperror.ErrValidation)
	}
	if input.ShopPhone == "" {
		return model.ShopSetting{}, apperror.BadRequest("shopPhone is required", apperror.ErrValidation)
	}
	if input.DepositAmount != nil && *input.DepositAmount < 0 {
		return model.ShopSetting{}, apperror.BadRequest("depositAmount must be greater than or equal to 0", apperror.ErrValidation)
	}

	setting, err := s.repo.Get()
	if err != nil {
		return model.ShopSetting{}, err
	}
	setting.ShopStatus = input.ShopStatus
	setting.OpenTime = input.OpenTime
	setting.CloseTime = input.CloseTime
	setting.ShopPhone = input.ShopPhone
	if input.PromptPayNumber != "" {
		setting.PromptPayNumber = input.PromptPayNumber
	}
	if input.AccountName != "" {
		setting.AccountName = input.AccountName
	}
	if input.BankName != "" {
		setting.BankName = input.BankName
	}
	if input.DepositAmount != nil {
		setting.DepositAmount = *input.DepositAmount
	}
	if err := s.repo.Save(&setting); err != nil {
		return model.ShopSetting{}, err
	}
	return setting, nil
}
