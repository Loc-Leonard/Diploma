package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/internal/admin"
	"github.com/Loc-Leonard/Diploma/internal/auth"
	"github.com/Loc-Leonard/Diploma/internal/config"
	"github.com/Loc-Leonard/Diploma/internal/db"
	"github.com/Loc-Leonard/Diploma/internal/models"
)

func main() {
	cfg := config.Load()
	database := db.MustConnect(cfg.DBDsn)

	// миграция через GORM (на всякий случай, если без docker-entrypoint-initdb.d)
	// if err := database.AutoMigrate(&models.User{}); err != nil {
	// 	log.Fatalf("auto migrate failed: %v", err)
	// }

	seedAdmin(database, cfg)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	auth.RegisterRoutes(r, database)
	admin.RegisterRoutes(r, database)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func seedAdmin(db *gorm.DB, cfg *config.Config) {
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
