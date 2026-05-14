package models

import "time"

type WorkReportStatus string

const (
	WorkReportStatusSubmitted WorkReportStatus = "SUBMITTED"
	WorkReportStatusVerified  WorkReportStatus = "VERIFIED"
	WorkReportStatusRejected  WorkReportStatus = "REJECTED"
)

type WorkReport struct {
	ID         uint             `gorm:"primaryKey" json:"id"`
	ObjectID   uint             `json:"object_id"`
	WorkItemID uint             `json:"work_item_id"`
	ForemanID  uint             `json:"foreman_id"`
	Date       time.Time        `json:"date"`
	Qty        float64          `json:"qty"`
	Status     WorkReportStatus `json:"status"`
	Comment    string           `json:"comment"`

	VerifiedBy      *uint      `json:"verified_by,omitempty"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty"`
	RejectionReason string     `json:"rejection_reason,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Поставка материала (упрощённо)
type MaterialDelivery struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ObjectID       uint      `json:"object_id"`
	WorkItemID     *uint     `json:"work_item_id,omitempty"`
	ForemanID      uint      `json:"foreman_id"`
	Date           time.Time `json:"date"`
	Material       string    `json:"material"`
	Qty            float64   `json:"qty"`
	Unit           string    `json:"unit"`
	DocumentNumber string    `json:"document_number"`
	Source         string    `json:"source"` // "MANUAL" | "CV_AUTO"
	CVConfidence   float64   `json:"cv_confidence"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// НЕ добавляйте Documents пока - GORM ругается
	// Documents     []MaterialDocument `gorm:"foreignKey:DeliveryID"`
}

type WorkItemStatus string

const (
	WorkItemStatusPlanned    WorkItemStatus = "PLANNED"
	WorkItemStatusInProgress WorkItemStatus = "IN_PROGRESS"
	WorkItemStatusDone       WorkItemStatus = "DONE"
	WorkItemStatusDelayed    WorkItemStatus = "DELAYED"
)

type WorkItem struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	ObjectID         uint           `json:"object_id"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	Unit             string         `json:"unit"`
	PlanQty          float64        `json:"plan_qty"`
	PlannedStartDate *time.Time     `json:"planned_start_date,omitempty"`
	PlannedEndDate   *time.Time     `json:"planned_end_date,omitempty"`
	ActualStartDate  *time.Time     `json:"actual_start_date,omitempty"`
	ActualEndDate    *time.Time     `json:"actual_end_date,omitempty"`
	SortOrder        int            `json:"sort_order"`
	Status           WorkItemStatus `json:"status"`
	DependsOnID      *uint          `json:"depends_on_id,omitempty"`
	Progress         float64        `json:"progress" gorm:"default:0"`
}
