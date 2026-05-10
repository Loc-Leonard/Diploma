package db

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/backend/internal/config"
	"github.com/Loc-Leonard/Diploma/backend/internal/models"
)

// SeedAdmin создает администратора если его еще нет
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

// SeedUsers создает тестовых пользователей с паролем "1" если их ещё нет
func SeedUsers(db *gorm.DB) {
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
			continue
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

// SeedSampleData создает тестовые объекты и инспекции
func SeedSampleData(db *gorm.DB) {
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
		inspections := []models.Inspection{
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
		if err := db.Create(&inspections).Error; err != nil {
			log.Printf("failed to create sample inspections: %v", err)
		}
	}
}
