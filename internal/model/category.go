package model

import "gorm.io/gorm"

type Category struct {
	gorm.Model

	CategoryID   string `gorm:"column:category_id;type:varchar(50);not null" json:"categoryId"`
	Name         string `gorm:"type:varchar(100);not null" json:"name"`
	Slug         string `gorm:"type:varchar(100);not null" json:"slug"`
	DisplayOrder int    `gorm:"default:0" json:"displayOrder"`
}

func (Category) TableName() string {
	return "category_dbs"
}
