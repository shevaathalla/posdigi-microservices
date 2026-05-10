package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID       string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Email    string `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password string `gorm:"type:varchar(255);not null" json:"-"`
}
