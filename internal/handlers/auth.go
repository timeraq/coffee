package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/timeraq/coffee/internal/middleware"
	"github.com/timeraq/coffee/internal/models"
)

type AuthHandler struct {
	DB *gorm.DB
}

type ShopRegisterRequest struct {
	Name          string `json:"name" binding:"required"`
	AdminEmail    string `json:"admin_email" binding:"required,email"`
	AdminPassword string `json:"admin_password" binding:"required,min=6"`
}

func (h *AuthHandler) RegisterShop(c *gin.Context) {
	var req ShopRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.CoffeeShop
	if err := h.DB.Where("admin_email = ?", req.AdminEmail).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "shop with this admin_email already exists"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	shop := models.CoffeeShop{
		ID:                uuid.New().String(),
		Name:              req.Name,
		AdminEmail:        req.AdminEmail,
		AdminPasswordHash: string(hash),
	}

	if err := h.DB.Create(&shop).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create shop"})
		return
	}

	token, err := middleware.GenerateToken(shop.ID, shop.AdminEmail, "", shop.ID, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":   token,
		"shop_id": shop.ID,
		"role":    "admin",
	})
}

type GuestRegisterRequest struct {
	Phone  string `json:"phone" binding:"required_without=Email"`
	Email  string `json:"email" binding:"required_without=Phone"`
	ShopID string `json:"shop_id" binding:"required"`
}

func (h *AuthHandler) RegisterGuest(c *gin.Context) {
	var req GuestRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.User
	q := h.DB.Where("coffee_shop_id = ?", req.ShopID)
	if req.Phone != "" {
		q = q.Where("phone = ?", req.Phone)
	} else {
		q = q.Where("email = ?", req.Email)
	}
	if err := q.First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}

	user := models.User{
		ID:           uuid.New().String(),
		Phone:        req.Phone,
		Email:        req.Email,
		CoffeeShopID: req.ShopID,
		Balance:      0,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, user.Phone, req.ShopID, "guest")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":   token,
		"user_id": user.ID,
		"role":    "guest",
		"user":    user,
	})
}

type GuestLoginRequest struct {
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	ShopID string `json:"shop_id" binding:"required"`
}

func (h *AuthHandler) LoginGuest(c *gin.Context) {
	var req GuestLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Phone == "" && req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone or email required"})
		return
	}

	var user models.User
	q := h.DB.Where("coffee_shop_id = ?", req.ShopID)
	if req.Phone != "" {
		q = q.Where("phone = ?", req.Phone)
	} else {
		q = q.Where("email = ?", req.Email)
	}
	if err := q.First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, user.Phone, req.ShopID, "guest")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"user_id": user.ID,
		"role":    "guest",
		"user":    user,
	})
}
