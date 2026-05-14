package db

import (
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/backend/internal/models"
)

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
