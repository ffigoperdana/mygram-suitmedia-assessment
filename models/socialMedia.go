package models

import (
	"github.com/asaskevich/govalidator"
	"gorm.io/gorm"
)

// SocialMedia represents the model for a SocialMedia
type SocialMedia struct {
	GormModel
	Name           string `gorm:"not null" json:"name" form:"name" valid:"required~Your name is required"`
	SocialMediaURL string `gorm:"not null" json:"social_media_url" form:"social_media_url" valid:"required~Your social media url is required,url~Invalid url format"`
	UserID         uint
	User           *User
}

func (s *SocialMedia) BeforeCreate(tx *gorm.DB) (err error) {
	_, errCreate := govalidator.ValidateStruct(s)

	if errCreate != nil {
		err = errCreate
		return
	}

	err = nil
	return
}

func (s *SocialMedia) BeforeUpdate(tx *gorm.DB) (err error) {
	_, errUpdate := govalidator.ValidateStruct(s)

	if errUpdate != nil {
		err = errUpdate
		return
	}

	err = nil
	return
}
