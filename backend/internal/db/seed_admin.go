package db

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/backend/internal/config"
	"github.com/Loc-Leonard/Diploma/backend/internal/models"
)

func SeedAdmin(db *gorm.DB, cfg *config.Config) {
	var count int64
	db.Model(&models.User{}).Where("email = ? AND role = ?", cfg.AdminEmail, models.RoleAdmin).Count(&count)
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt error: %v", err)
	}

	email := cfg.AdminEmail
	adminUser := models.User{
		FullName:           cfg.AdminName,
		Email:              &email,
		Role:               models.RoleAdmin,
		PasswordHash:       string(hash),
		MustChangePassword: false,
	}

	if err := db.Create(&adminUser).Error; err != nil {
		log.Fatalf("failed to create admin: %v", err)
	}
}
