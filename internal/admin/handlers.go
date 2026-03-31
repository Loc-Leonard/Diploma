package admin

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/internal/auth"
	"github.com/Loc-Leonard/Diploma/internal/models"
)

type Handler struct {
	db *gorm.DB
}

const tempPasswordLength = 10

func generateTempPassword() (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var b = make([]byte, tempPasswordLength)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[n.Int64()]
	}
	return string(b), nil
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	h := &Handler{db: db}

	gr := r.Group("/admin")
	gr.Use(auth.AuthMiddleware(), auth.AdminOnly())
	{
		gr.POST("/users", h.CreateUser)
		gr.GET("/users", h.ListUsers)
	}
}

type CreateUserRequest struct {
	FullName string      `json:"full_name" binding:"required"`
	Email    *string     `json:"email"`
	Phone    *string     `json:"phone"`
	Role     models.Role `json:"role" binding:"required"`
}

type UserListItem struct {
	ID        uint        `json:"id"`
	FullName  string      `json:"full_name"`
	Email     *string     `json:"email"`
	Phone     *string     `json:"phone"`
	Role      models.Role `json:"role"`
	Status    string      `json:"status"`     // пока захардкодим ACTIVE
	LastLogin *time.Time  `json:"last_login"` // если появится поле в модели
}

func (h *Handler) ListUsers(c *gin.Context) {
	var users []models.User
	if err := h.db.Order("id ASC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	resp := make([]UserListItem, 0, len(users))
	for _, u := range users {
		item := UserListItem{
			ID:       u.ID,
			FullName: u.FullName,
			Email:    u.Email,
			Phone:    u.Phone,
			Role:     u.Role,
			Status:   "ACTIVE", // позже можно добавить реальное поле в модель
		}
		resp = append(resp, item)
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tempPassword, err := generateTempPassword()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password generate error"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
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
		"id":            user.ID,
		"full_name":     user.FullName,
		"email":         user.Email,
		"phone":         user.Phone,
		"role":          user.Role,
		"temp_password": tempPassword, // ← показываем админу
	})
}

func HandlerForTest(db *gorm.DB) *Handler {
	return &Handler{db: db}
}
