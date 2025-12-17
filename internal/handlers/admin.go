package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/timeraq/coffee/internal/models"
)

type AdminHandler struct {
	DB *gorm.DB
}

func (h *AdminHandler) GetDashboard(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var totalGuests int64
	h.DB.Model(&models.User{}).
		Where("coffee_shop_id = ?", shopID).
		Count(&totalGuests)

	var totalRevenue float64
	h.DB.Model(&models.Purchase{}).
		Where("shop_id = ?", shopID).
		Select("COALESCE(SUM(amount),0)").
		Scan(&totalRevenue)

	var churnRisk int64
	h.DB.Model(&models.User{}).
		Where("coffee_shop_id = ? AND (last_visit IS NULL OR last_visit < ?)",
			shopID, time.Now().AddDate(0, 0, -14)).
		Count(&churnRisk)

	type TopGuest struct {
		UserID     string  `json:"user_id"`
		Phone      string  `json:"phone"`
		Email      string  `json:"email"`
		TotalSpent float64 `json:"total_spent"`
		Visits     int64   `json:"visits"`
	}

	var rows []TopGuest
	h.DB.Model(&models.Purchase{}).
		Select("user_id, COUNT(*) as visits, SUM(amount) as total_spent").
		Where("shop_id = ?", shopID).
		Group("user_id").
		Order("total_spent DESC").
		Limit(10).
		Scan(&rows)

	for i, g := range rows {
		var u models.User
		if err := h.DB.Where("id = ?", g.UserID).First(&u).Error; err == nil {
			rows[i].Phone = u.Phone
			rows[i].Email = u.Email
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_guests":     totalGuests,
		"total_revenue":    totalRevenue,
		"churn_risk_count": churnRisk,
		"top_guests":       rows,
	})
}

func (h *AdminHandler) GetChurnRiskUsers(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var users []models.User
	h.DB.Where("coffee_shop_id = ? AND (last_visit IS NULL OR last_visit < ?)",
		shopID, time.Now().AddDate(0, 0, -14)).
		Find(&users)

	c.JSON(http.StatusOK, gin.H{"users": users})
}

type UpdateShopSettingsRequest struct {
	PointsPer100Rub     *int    `json:"points_per_100_rub"`
	FreeCoeffeeAtPoints *int    `json:"free_coffee_at_points"`
	Color               *string `json:"color"`
}

func (h *AdminHandler) UpdateShopSettings(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var req UpdateShopSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.PointsPer100Rub != nil {
		updates["points_per_100_rub"] = *req.PointsPer100Rub
	}
	if req.FreeCoeffeeAtPoints != nil {
		updates["free_coffee_at_points"] = *req.FreeCoeffeeAtPoints
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	if err := h.DB.Model(&models.CoffeeShop{}).
		Where("id = ?", shopID).
		Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "settings updated"})
}
