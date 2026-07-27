package model

import "gorm.io/gorm"

type NailTechnician struct {
	gorm.Model

	TechnicianID    string `gorm:"column:technician_id;type:varchar(50)" json:"technicianId"`
	TechnicianName  string `gorm:"type:varchar(255);not null" json:"technicianName"`
	Phone           string `gorm:"type:varchar(50)" json:"phone,omitempty"`
	ExperienceYears int    `gorm:"default:0" json:"experienceYears"`
	Specialty       string `gorm:"type:varchar(255)" json:"specialty,omitempty"`
	ProfileImg      string `gorm:"type:text" json:"profileImg,omitempty"`
	AvatarURL       string `gorm:"type:text" json:"avatarUrl,omitempty"`
	Active          bool   `gorm:"default:true" json:"active"`
	Bio             string `gorm:"type:text" json:"bio,omitempty"`
}

func (NailTechnician) TableName() string {
	return "nail_technician_dbs"
}
