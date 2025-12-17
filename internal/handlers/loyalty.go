package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"

	"github.com/timeraq/coffee/internal/models"
	"github.com/timeraq/coffee/internal/services"
)

type LoyaltyHandler struct {
	DB *gorm.DB
	PS *services.PointsService
}

type AddPointsRequest struct {
	CheckID string   `json:"check_id" binding:"required"`
	Amount  float64  `json:"amount" binding:"required"`
	Items   []string `json:"items,omitempty"`
}

// Гость добавляет чек → начисляем баллы
func (h *LoyaltyHandler) AddPoints(c *gin.Context) {
	userID := c.GetString("user_id")
	shopID := c.GetString("shop_id")

	var req AddPointsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка дубликата чека
	var existing models.Purchase
	if err := h.DB.Where("check_id = ? AND user_id = ?", req.CheckID, userID).
		First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "check already processed"})
		return
	}

	// Правила из кофейни
	var shop models.CoffeeShop
	if err := h.DB.Where("id = ?", shopID).First(&shop).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shop not found"})
		return
	}

	points := h.PS.CalculatePoints(req.Amount, shop.PointsPer100Rub)

	purchase := models.Purchase{
		ID:      uuid.New().String(),
		UserID:  userID,
		ShopID:  shopID,
		CheckID: req.CheckID,
		Amount:  req.Amount,
		Points:  points,
		Items:   models.JSONArray(req.Items),
		Status:  "confirmed",
	}

	if err := h.DB.Create(&purchase).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save purchase"})
		return
	}

	_ = h.PS.AddPoints(userID, points)
	h.PS.UpdateLastVisit(userID)

	var user models.User
	h.DB.Where("id = ?", userID).First(&user)

	c.JSON(http.StatusOK, gin.H{
		"purchase":       purchase,
		"new_balance":    user.Balance,
		"to_free_coffee": shop.FreeCoeffeeAtPoints - user.Balance,
	})
}

func (h *LoyaltyHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var user models.User
	if err := h.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var shop models.CoffeeShop
	h.DB.Where("id = ?", user.CoffeeShopID).First(&shop)

	c.JSON(http.StatusOK, gin.H{
		"user":               user,
		"free_coffee_points": shop.FreeCoeffeeAtPoints,
		"progress_to_free":   shop.FreeCoeffeeAtPoints - user.Balance,
	})
}

func (h *LoyaltyHandler) GetHistory(c *gin.Context) {
	userID := c.GetString("user_id")

	var purchases []models.Purchase
	if err := h.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(100).
		Find(&purchases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"purchases": purchases})
}

type RedeemRequest struct {
	Type string `json:"type" binding:"required"` // "free_coffee"
}

func (h *LoyaltyHandler) Redeem(c *gin.Context) {
	userID := c.GetString("user_id")
	shopID := c.GetString("shop_id")

	var req RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var shop models.CoffeeShop
	if err := h.DB.Where("id = ?", shopID).First(&shop).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shop not found"})
		return
	}

	if err := h.PS.RedeemFreeCoffee(userID, shop.FreeCoeffeeAtPoints); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient points"})
		return
	}

	h.PS.UpdateLastVisit(userID)

	c.JSON(http.StatusOK, gin.H{"message": "free coffee redeemed"})
}
