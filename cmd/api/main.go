package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/timeraq/coffee/internal/db"
	"github.com/timeraq/coffee/internal/handlers"
	"github.com/timeraq/coffee/internal/middleware"
	"github.com/timeraq/coffee/internal/services"
)

func main() {
	_ = godotenv.Load()

	log.Println("🚀 Starting Coffee Loyalty Backend...")

	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}

	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("❌ Migrations failed: %v", err)
	}

	pointsService := &services.PointsService{DB: database}

	authHandler := &handlers.AuthHandler{DB: database}
	loyaltyHandler := &handlers.LoyaltyHandler{DB: database, PS: pointsService}
	adminHandler := &handlers.AdminHandler{DB: database}

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// AUTH
	auth := r.Group("/api/auth")
	{
		auth.POST("/register-shop", authHandler.RegisterShop)
		auth.POST("/register-guest", authHandler.RegisterGuest)
		auth.POST("/login-guest", authHandler.LoginGuest)
	}

	// LOYALTY (гости)
	loyalty := r.Group("/api/loyalty")
	loyalty.Use(middleware.AuthMiddleware())
	{
		loyalty.GET("/profile", loyaltyHandler.GetProfile)
		loyalty.GET("/history", loyaltyHandler.GetHistory)
		loyalty.POST("/add-points", loyaltyHandler.AddPoints)
		loyalty.POST("/redeem", loyaltyHandler.Redeem)
	}

	// ADMIN (кофейни)
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.GET("/dashboard", adminHandler.GetDashboard)
		admin.GET("/churn-risk", adminHandler.GetChurnRiskUsers)
		admin.PUT("/settings", adminHandler.UpdateShopSettings)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("✅ Backend is ready on port", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Server stopped: %v", err)
	}
}
