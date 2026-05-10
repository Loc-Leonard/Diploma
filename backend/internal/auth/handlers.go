package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/backend/internal/models"
)

// Route paths
const (
	RouteAuthLogin          = "/auth/login"
	RouteAuthMe             = "/auth/me"
	RouteAuthChangePassword = "/auth/change-password"
)

type Handler struct {
	db *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	h := &Handler{db: db}

	gr := r.Group("/auth")
	{
		gr.POST("/login", h.Login)
		gr.GET("/me", AuthMiddleware(), h.Me)
		gr.POST("/change-password", AuthMiddleware(), h.ChangePassword)
	}
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token              string         `json:"token"`
	User               models.UserDTO `json:"user"`
	MustChangePassword bool           `json:"must_change_password"`
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("email = ? OR phone = ?", req.Login, req.Login).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid credentials"})
		return
	}

	token, err := GenerateToken(user.ID, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "token error"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User: models.UserDTO{
			ID:       user.ID,
			FullName: user.FullName,
			Role:     user.Role,
		},
		MustChangePassword: user.MustChangePassword,
	})
}

func (h *Handler) Me(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "user not found"})
		return
	}

	c.JSON(http.StatusOK, models.UserDTO{
		ID:       user.ID,
		FullName: user.FullName,
		Role:     user.Role,
	})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *Handler) ChangePassword(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "user not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "old password incorrect"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "hash error"})
		return
	}

	user.PasswordHash = string(hash)
	user.MustChangePassword = false

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "save error"})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Status: "ok"})
}
