package model

import "gorm.io/gorm"

type Service struct {
	gorm.Model

	ServiceID    string `gorm:"column:service_id;type:varchar(50);not null" json:"serviceId"`
	ServiceName  string `gorm:"type:varchar(255);not null" json:"name"`
	ServicePrice int    `gorm:"not null" json:"price"`
	Duration     int    `gorm:"not null" json:"duration"`
	Category     string `gorm:"type:varchar(100)" json:"category"`
	CategoryID   *uint  `gorm:"column:category_id" json:"categoryId,omitempty"`
	ImageURL     string `gorm:"type:text" json:"imageUrl,omitempty"`
	Img          string `gorm:"column:img;type:text" json:"img,omitempty"`
	ServiceImg   string `gorm:"type:text" json:"-"`
	Popular      bool   `gorm:"default:false" json:"popular"`
	Description  string `gorm:"type:text" json:"description,omitempty"`
}

func (Service) TableName() string {
	return "service_dbs"
}
