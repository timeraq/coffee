package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type CoffeeShop struct {
	ID                  string    `gorm:"primaryKey" json:"id"`
	Name                string    `json:"name"`
	AdminEmail          string    `gorm:"uniqueIndex" json:"admin_email"`
	AdminPasswordHash   string    `json:"-"`
	TelegramChatID      int64     `json:"telegram_chat_id,omitempty"`
	PointsPer100Rub     int       `json:"points_per_100_rub" gorm:"default:1"`
	FreeCoeffeeAtPoints int       `json:"free_coffee_at_points" gorm:"default:600"`
	CreatedAt           time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CoffeeShop) TableName() string {
	return "coffee_shops"
}

func (c *CoffeeShop) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
