package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	Phone        string     `gorm:"uniqueIndex;index" json:"phone,omitempty"`
	Email        string     `gorm:"uniqueIndex;index" json:"email,omitempty"`
	TelegramID   int64      `gorm:"uniqueIndex;index" json:"telegram_id,omitempty"`
	Balance      int        `json:"balance" gorm:"default:0"`
	CoffeeShopID string     `json:"coffee_shop_id"`
	LastVisit    *time.Time `json:"last_visit"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}
