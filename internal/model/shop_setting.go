package model

import "gorm.io/gorm"

type ShopSetting struct {
	gorm.Model

	ShopStatus      string  `gorm:"type:varchar(10);not null;default:open;check:shop_status IN ('open','closed')" json:"shopStatus"`
	OpenTime        string  `gorm:"type:varchar(5);not null;default:10:00" json:"openTime"`
	CloseTime       string  `gorm:"type:varchar(5);not null;default:20:00" json:"closeTime"`
	ShopPhone       string  `gorm:"type:varchar(50);not null;default:02-123-4567" json:"shopPhone"`
	PromptPayNumber string  `gorm:"type:varchar(20);not null;default:0812345678" json:"promptPayNumber"`
	AccountName     string  `gorm:"type:varchar(100);not null;default:ร้าน Nailly Nail Salon" json:"accountName"`
	BankName        string  `gorm:"type:varchar(50);not null;default:ธนาคารกสิกรไทย" json:"bankName"`
	DepositAmount   float64 `gorm:"type:numeric(10,2);not null;default:200;check:deposit_amount >= 0" json:"depositAmount"`
	QrCodeUrl       string  `gorm:"type:text" json:"qrCodeUrl"`
}

func (ShopSetting) TableName() string {
	return "shop_settings"
}

func DefaultShopSetting() ShopSetting {
	return ShopSetting{
		Model:           gorm.Model{ID: 1},
		ShopStatus:      "open",
		OpenTime:        "10:00",
		CloseTime:       "20:00",
		ShopPhone:       "02-123-4567",
		PromptPayNumber: "0812345678",
		AccountName:     "ร้าน Nailly Nail Salon",
		BankName:        "ธนาคารกสิกรไทย",
		DepositAmount:   200,
		QrCodeUrl:       "",
	}
}
