package models

// UserDTO - базовый DTO пользователя
type UserDTO struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
	Role     Role   `json:"role"`
}

// SimpleUserDTO - упрощенный DTO пользователя для списков
type SimpleUserDTO struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
	City     string `json:"city,omitempty"`
}

// ErrorResponse - единый формат ответа при ошибке
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse - единый формат ответа при успехе
type SuccessResponse struct {
	Status string `json:"status"`
}

// IDResponse - ответ с ID созданного ресурса
type IDResponse struct {
	ID uint `json:"id"`
}
