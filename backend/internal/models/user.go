package models

import "time"

type Role string

const (
	RoleAdmin     Role = "ADMIN"
	RoleExecutor  Role = "EXECUTOR" //вырезать потом!
	RoleCustomer  Role = "CUSTOMER"
	RoleInspector Role = "INSPECTOR"
	RoleForeman   Role = "FOREMAN"
)

type User struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	FullName           string    `json:"full_name"`
	Email              *string   `gorm:"uniqueIndex"` // без имени
	Phone              *string   `gorm:"uniqueIndex"` // без имени
	Role               Role      `json:"role"`
	PasswordHash       string    `json:"-"`
	MustChangePassword bool      `json:"must_change_password"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
