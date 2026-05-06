package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/backend/internal/admin"
	"github.com/Loc-Leonard/Diploma/backend/internal/auth"
	"github.com/Loc-Leonard/Diploma/backend/internal/config"
	"github.com/Loc-Leonard/Diploma/backend/internal/customer"
	"github.com/Loc-Leonard/Diploma/backend/internal/cv"
	"github.com/Loc-Leonard/Diploma/backend/internal/db"
	"github.com/Loc-Leonard/Diploma/backend/internal/foreman"
	"github.com/Loc-Leonard/Diploma/backend/internal/inspector"
	"github.com/Loc-Leonard/Diploma/backend/internal/models"
)

func main() {
	cfg := config.Load()
	database := db.MustConnect(cfg.DBDsn)
	cvProcessor := cv.HTTPProcessor{
		BaseURL: cfg.CVServiceURL,
	}

	if err := database.AutoMigrate(
		&models.User{},
		&models.Object{},
		&models.WorkItem{},
		&models.WorkReport{},
		&models.MaterialDelivery{},
		&models.MaterialDocument{},
		&models.Inspection{},
	); err != nil {
		log.Printf("auto migrate failed: %v", err)
	}

	seedAdmin(database, cfg)
	seedUsers(database)      // ← сначала создаём всех юзеров
	seedSampleData(database) // ← потом sample data, которая их ищет

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
	foreman.RegisterRoutes(r, database, cvProcessor, cfg.StorageRoot)
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

// seedUsers создаёт тестовых пользователей с паролем "1" если их ещё нет.
// Идемпотентна — при повторном запуске дубли не создаются.
func seedUsers(db *gorm.DB) {
	hash, err := bcrypt.GenerateFromPassword([]byte("1"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("seed: cannot hash password:", err)
	}
	h := string(hash)

	users := []struct {
		email    string
		fullName string
		role     models.Role
	}{
		{"customer@test.ru", "Заказчик Тест", models.RoleCustomer},
		{"foreman@test.ru", "Прораб Тест", models.RoleForeman},
		{"inspector@test.ru", "Инспектор Тест", models.RoleInspector},
	}

	for _, u := range users {
		var existing models.User
		if err := db.Where("email = ?", u.email).First(&existing).Error; err == nil {
			continue // уже есть — пропускаем
		}

		email := u.email
		newUser := models.User{
			FullName:           u.fullName,
			Email:              &email,
			Role:               u.role,
			PasswordHash:       h,
			MustChangePassword: false,
		}

		if err := db.Create(&newUser).Error; err != nil {
			log.Printf("seed: cannot create %s: %v", u.fullName, err)
		} else {
			log.Printf("seed: created %s / %s / pass: 1", u.fullName, u.email)
		}
	}
}

func seedSampleData(db *gorm.DB) {
	var customerUser models.User
	if err := db.Where("role = ?", models.RoleCustomer).First(&customerUser).Error; err != nil {
		log.Println("no customer found for seeding sample data")
		return
	}

	var foremanUser models.User
	if err := db.Where("role = ?", models.RoleForeman).First(&foremanUser).Error; err != nil {
		log.Println("no foreman found for seeding sample data")
		return
	}

	var inspectorUser models.User
	if err := db.Where("role = ?", models.RoleInspector).First(&inspectorUser).Error; err != nil {
		log.Println("no inspector found for seeding sample inspections")
	}

	var count int64
	db.Model(&models.Object{}).
		Where("name = ? AND customer_control_user_id = ?", "Объект #1 «Парк Победы»", customerUser.ID).
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
		CustomerControlUserID: customerUser.ID,
		ForemanUserID:         foremanUser.ID,
		InspectorUserID:       inspectorUser.ID,
	}

	if err := db.Create(&obj).Error; err != nil {
		log.Printf("failed to create sample object: %v", err)
		return
	}

	if inspectorUser.ID != 0 {
		now := time.Now()
		insp := []models.Inspection{
			{
				ObjectID:    obj.ID,
				InspectorID: inspectorUser.ID,
				Status:      models.InspectionStatusPlanned,
				PlannedAt:   now.AddDate(0, 0, 3),
				IssuesOpen:  0,
			},
			{
				ObjectID:    obj.ID,
				InspectorID: inspectorUser.ID,
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
