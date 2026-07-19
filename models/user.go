package models

import (
	"finalproject/helpers"
	"time"

	"github.com/asaskevich/govalidator"
	"gorm.io/gorm"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"

	UserStatusActive = "active"
	UserStatusBanned = "banned"
)

// User represents the model for an User
type User struct {
	GormModel
	Username          string             `gorm:"not null;uniqueIndex" json:"username" form:"username" valid:"required~Your username is required"`
	Email             string             `gorm:"not null;uniqueIndex" json:"email" form:"email" valid:"required~Your email is required,email~Invalid email format"`
	Password          string             `gorm:"not null" json:"-" form:"password" valid:"required~Your password is required,minstringlength(6)~Password has to have a minimum length of 6 characters"`
	Age               int                `gorm:"not null" json:"age" form:"age" valid:"required~Your age is required,range(9|100)~Age has to be above 8 years old"`
	Role              string             `gorm:"not null;default:user" json:"role" form:"role"`
	Status            string             `gorm:"not null;default:active" json:"status" form:"status"`
	BannedAt          *time.Time         `json:"banned_at,omitempty"`
	BanReason         string             `json:"ban_reason,omitempty" form:"ban_reason"`
	LastLoginAt       *time.Time         `json:"last_login_at,omitempty"`
	LastSeenAt        *time.Time         `json:"last_seen_at,omitempty"`
	Photos            []Photo            `gorm:"constraint:OnUpdate:CASCADE,onDelete:SET NULL;" json:"photos"`
	Comments          []Comment          `gorm:"constraint:OnUpdate:CASCADE,onDelete:SET NULL;" json:"comments"`
	SocialMedias      []SocialMedia      `gorm:"constraint:OnUpdate:CASCADE,onDelete:SET NULL;" json:"social_medias"`
	PushSubscriptions []PushSubscription `gorm:"constraint:OnUpdate:CASCADE,onDelete:CASCADE;" json:"push_subscriptions"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Role == "" {
		u.Role = RoleUser
	}
	if u.Status == "" {
		u.Status = UserStatusActive
	}

	_, errCreate := govalidator.ValidateStruct(u)

	if errCreate != nil {
		err = errCreate
		return
	}

	u.Password = helpers.HashPass(u.Password)
	err = nil
	return
}

func IsValidUserRole(role string) bool {
	return role == RoleUser || role == RoleAdmin
}

func IsValidUserStatus(status string) bool {
	return status == UserStatusActive || status == UserStatusBanned
}
