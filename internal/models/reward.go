package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Reward struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	UserID       string     `gorm:"index" json:"user_id"`
	ShopID       string     `gorm:"index" json:"shop_id"`
	Type         string     `json:"type"`
	Title        string     `json:"title"`
	Description  string     `json:"description,omitempty"`
	PointsReward int        `json:"points_reward"`
	UnlockedAt   *time.Time `json:"unlocked_at"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

func (Reward) TableName() string {
	return "rewards"
}

func (r *Reward) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}
