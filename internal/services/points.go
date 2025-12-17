package services

import (
	"time"

	"github.com/timeraq/coffee/internal/models"
	"gorm.io/gorm"
)

type PointsService struct {
	DB *gorm.DB
}

// Счёт баллов: amount в рублях, pointsPer100Rub — правило кофейни (например 1 балл за 100р)
func (ps *PointsService) CalculatePoints(amount float64, pointsPer100Rub int) int {
	points := int(amount/100.0) * pointsPer100Rub
	if points < 1 {
		points = 1
	}
	return points
}

func (ps *PointsService) AddPoints(userID string, points int) error {
	return ps.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("balance", gorm.Expr("balance + ?", points)).Error
}

func (ps *PointsService) RedeemFreeCoffee(userID string, required int) error {
	return ps.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}
		if user.Balance < required {
			return gorm.ErrRecordNotFound
		}
		return tx.Model(&user).
			Update("balance", gorm.Expr("balance - ?", required)).Error
	})
}

// Пример: обновить last_visit при покупке
func (ps *PointsService) UpdateLastVisit(userID string) {
	ps.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("last_visit", time.Now())
}
