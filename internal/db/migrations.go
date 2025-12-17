package db

import (
	"github.com/timeraq/coffee/internal/models"
	"gorm.io/gorm"
	"log"
)

func RunMigrations(db *gorm.DB) error {
	log.Println("🔄 Running migrations...")

	if err := db.AutoMigrate(
		&models.CoffeeShop{},
		&models.User{},
		&models.Purchase{},
		&models.Reward{},
	); err != nil {
		return err
	}

	log.Println("✅ Migrations completed!")
	return nil
}
