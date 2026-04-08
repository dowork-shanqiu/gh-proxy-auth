package models

import (
	"time"
)

type DownloadLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TokenID   uint      `gorm:"index;not null" json:"token_id"`
	Token     Token     `gorm:"foreignKey:TokenID" json:"token,omitempty"`
	URL       string    `gorm:"size:2048;not null" json:"url"`
	IP        string    `gorm:"size:64" json:"ip"`
	UserAgent string    `gorm:"size:512" json:"user_agent"`
}
