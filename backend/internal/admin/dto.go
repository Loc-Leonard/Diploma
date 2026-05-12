package admin

import (
	"time"

	"github.com/Loc-Leonard/Diploma/backend/internal/models"
)

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
	Status    string      `json:"status"`
	LastLogin *time.Time  `json:"last_login"`
}
