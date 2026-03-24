package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/internal/auth"
	"github.com/Loc-Leonard/Diploma/internal/models"
)

type Handler struct {
	db *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	h := &Handler{db: db}

	gr := r.Group("/admin")
	gr.Use(auth.AuthMiddleware(), auth.AdminOnly())
	{
		gr.POST("/users", h.CreateUser)
	}
}

type CreateUserRequest struct {
	FullName string      `json:"full_name" binding:"required"`
	Email    *string     `json:"email"`
	Phone    *string     `json:"phone"`
	Role     models.Role `json:"role" binding:"required"`
	Password string      `json:"password" binding:"required"`
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash error"})
		return
	}

	user := models.User{
		FullName:           req.FullName,
		Email:              req.Email,
		Phone:              req.Phone,
		Role:               req.Role,
		PasswordHash:       string(hash),
		MustChangePassword: true,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        user.ID,
		"full_name": user.FullName,
		"email":     user.Email,
		"phone":     user.Phone,
		"role":      user.Role,
	})
}
