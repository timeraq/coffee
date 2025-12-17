package models

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Purchase struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index" json:"user_id"`
	ShopID    string    `gorm:"index" json:"shop_id"`
	CheckID   string    `gorm:"index" json:"check_id"`
	Amount    float64   `json:"amount"`
	Points    int       `json:"points"`
	Items     JSONArray `json:"items,omitempty"`
	Status    string    `json:"status" gorm:"default:'confirmed'"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type JSONArray []string

func (a JSONArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *JSONArray) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &a)
}

func (Purchase) TableName() string {
	return "purchases"
}

func (p *Purchase) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
