package db

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/backend/internal/models"
)

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
