package models

import "time"

type PushSubscription struct {
	GormModel
	UserID     uint       `gorm:"not null;index" json:"user_id"`
	Endpoint   string     `gorm:"not null;size:2048;uniqueIndex" json:"endpoint"`
	P256DH     string     `gorm:"not null;size:512" json:"p256dh"`
	Auth       string     `gorm:"not null;size:256" json:"auth"`
	UserAgent  string     `gorm:"size:512" json:"user_agent,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	User       *User      `json:"user,omitempty"`
}
