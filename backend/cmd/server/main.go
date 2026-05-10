package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

	// Очищаем старые ограничения перед миграцией
	cleanupConstraints(database)

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

	// Сидирование данных
	db.SeedAdmin(database, cfg)
	db.SeedUsers(database)
	db.SeedSampleData(database)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Регистрация роутов
	auth.RegisterRoutes(r, database)
	admin.RegisterRoutes(r, database)
	customer.RegisterRoutes(r, database)
	foreman.RegisterRoutes(r, database, cvProcessor, cfg.StorageRoot)
	inspector.RegisterRoutes(r, database)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

// cleanupConstraints удаляет старые ограничения перед миграцией
func cleanupConstraints(database *gorm.DB) {
	database.Exec(`ALTER TABLE "users" DROP CONSTRAINT IF EXISTS "uni_users_email"`)
	database.Exec(`ALTER TABLE "users" DROP CONSTRAINT IF EXISTS "uni_users_phone"`)
	database.Exec(`DROP INDEX IF EXISTS "uni_users_email"`)
	database.Exec(`DROP INDEX IF EXISTS "uni_users_phone"`)

	database.Exec(`DROP TABLE IF EXISTS "objects" CASCADE`)
	database.Exec(`DROP TABLE IF EXISTS "work_items" CASCADE`)
	database.Exec(`DROP TABLE IF EXISTS "work_reports" CASCADE`)
	database.Exec(`DROP TABLE IF EXISTS "material_deliveries" CASCADE`)
	database.Exec(`DROP TABLE IF EXISTS "material_documents" CASCADE`)
	database.Exec(`DROP TABLE IF EXISTS "inspections" CASCADE`)
}
