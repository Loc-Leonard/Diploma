package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/internal/admin"
	"github.com/Loc-Leonard/Diploma/internal/auth"
	"github.com/Loc-Leonard/Diploma/internal/config"
	"github.com/Loc-Leonard/Diploma/internal/customer"
	"github.com/Loc-Leonard/Diploma/internal/db"
	"github.com/Loc-Leonard/Diploma/internal/foreman"
	"github.com/Loc-Leonard/Diploma/internal/inspector"
	"github.com/Loc-Leonard/Diploma/internal/models"
)

func main() {
	cfg := config.Load()
	database := db.MustConnect(cfg.DBDsn)

	// миграция через GORM (на всякий случай, если без docker-entrypoint-initdb.d)
	if err := database.AutoMigrate(
		&models.User{},
		&models.Object{},
		&models.WorkItem{},
		&models.WorkReport{},
		&models.MaterialDelivery{},
		&models.Inspection{},
	); err != nil {
		log.Printf("auto migrate failed: %v", err)
	}

	seedAdmin(database, cfg)
	seedSampleData(database)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	auth.RegisterRoutes(r, database)
	admin.RegisterRoutes(r, database)
	customer.RegisterRoutes(r, database)
	foreman.RegisterRoutes(r, database)
	inspector.RegisterRoutes(r, database)

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

func seedSampleData(db *gorm.DB) {
	// 1. ищем любого заказчика
	var customer models.User
	if err := db.Where("role = ?", models.RoleCustomer).First(&customer).Error; err != nil {
		log.Println("no customer found for seeding sample data")
		return
	}

	// 2. ищем любого прораба
	var foreman models.User
	if err := db.Where("role = ?", models.RoleForeman).First(&foreman).Error; err != nil {
		log.Println("no foreman found for seeding sample data")
		return
	}
	var inspector models.User
	if err := db.Where("role = ?", models.RoleInspector).First(&inspector).Error; err != nil {
		log.Println("no inspector found for seeding sample inpections")
	}

	// 3. проверяем, нет ли уже объекта
	var count int64
	db.Model(&models.Object{}).
		Where("name = ? AND customer_control_user_id = ?", "Объект #1 «Парк Победы»", customer.ID).
		Count(&count)
	if count > 0 {
		return
	}

	obj := models.Object{
		Name:                  "Объект #1 «Парк Победы»",
		Address:               "Москва, ул. Примерная, д. 1",
		City:                  "Москва",
		Description:           "Тестовый объект для дашборда",
		Status:                models.ObjectStatusActive,
		Lat:                   55.751244,
		Lng:                   37.618423,
		CustomerControlUserID: customer.ID,
		ForemanUserID:         foreman.ID,
		InspectorUserID:       inspector.ID,
	}

	if err := db.Create(&obj).Error; err != nil {
		log.Printf("failed to create sample object: %v", err)
	}

	if inspector.ID != 0 {
		now := time.Now()
		insp := []models.Inspection{
			{
				ObjectID:    obj.ID,
				InspectorID: inspector.ID,
				Status:      models.InspectionStatusPlanned,
				PlannedAt:   now.AddDate(0, 0, 3),
				IssuesOpen:  0,
			},
			{
				ObjectID:    obj.ID,
				InspectorID: inspector.ID,
				Status:      models.InspectionStatusOverdue,
				PlannedAt:   now.AddDate(0, 0, -2),
				IssuesOpen:  3,
			},
		}
		if err := db.Create(&insp).Error; err != nil {
			log.Printf("failed to create sample inspections: %v", err)
		}
	}
}
