package handler

import (
	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/dto"
	"nailly-back-end/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ShopSettingHandler struct {
	service *service.ShopSettingService
}

func NewShopSettingHandler(shopSettingService *service.ShopSettingService) *ShopSettingHandler {
	return &ShopSettingHandler{service: shopSettingService}
}

func (h *ShopSettingHandler) GetSettings(c *gin.Context) {
	setting, err := h.service.GetSettings()
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToShopSettingResponse(setting))
}

func (h *ShopSettingHandler) UpdateSettings(c *gin.Context) {
	var request dto.UpdateShopSettingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, apperror.BadRequest("invalid request body", err))
		return
	}
	setting, err := h.service.UpdateSettings(service.UpdateShopSettingInput{
		ShopStatus:      request.ShopStatus,
		OpenTime:        request.OpenTime,
		CloseTime:       request.CloseTime,
		ShopPhone:       request.ShopPhone,
		PromptPayNumber: request.PromptPayNumber,
		AccountName:     request.AccountName,
		BankName:        request.BankName,
		DepositAmount:   request.DepositAmount,
		QrCodeUrl:       request.QrCodeUrl,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToShopSettingResponse(setting))
}
