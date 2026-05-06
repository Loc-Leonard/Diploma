package models

import "time"

type ObjectStatus string

const (
	ObjectStatusPlanned                      ObjectStatus = "PLANNED"
	ObjectStatusWaitingInspectorConfirmation ObjectStatus = "WAITING_INSPECTOR_CONFIRMATION"
	ObjectStatusActive                       ObjectStatus = "ACTIVE"
	ObjectStatusFinished                     ObjectStatus = "FINISHED"
)

type Object struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Description string `json:"description"`

	Status ObjectStatus `json:"status"`

	// Координаты для карты (потом добавим polygon)
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`

	CustomerControlUserID uint `json:"customer_control_user_id"`
	ForemanUserID         uint `json:"foreman_user_id"`
	InspectorUserID       uint `json:"inspector_user_id"`

	PlannedStartDate *time.Time `json:"planned_start_date"`
	PlannedEndDate   *time.Time `json:"planned_end_date"`
	ActualStartDate  *time.Time `json:"actual_start_date"`
	ActualEndDate    *time.Time `json:"actual_end_date"`

	InitChecklistJSON string `json:"init_checklist_json"`
	InitActFilePath   string `json:"init_act_file_path"`

	ActivationRejectReason string     `json:"activation_reject_reason"`
	ActivationReviewedAt   *time.Time `json:"activation_reviewed_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Progress float64 `json:"progress" gorm:"default:0"`
}
