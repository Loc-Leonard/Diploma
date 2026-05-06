package models

import "time"

type WorkItem struct {
	ID       uint    `gorm:"primaryKey" json:"id"`
	ObjectID uint    `json:"object_id"`
	Name     string  `json:"name"`
	Unit     string  `json:"unit"`     // м2, шт — можно оставить пустым
	PlanQty  float64 `json:"plan_qty"` // плановый объем
}

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
}

// Поставка материала (упрощённо)
type MaterialDelivery struct {
	ID             uint               `gorm:"primaryKey" json:"id"`
	ObjectID       uint               `json:"object_id"`
	WorkItemID     *uint              `json:"work_item_id,omitempty"`
	ForemanID      uint               `json:"foreman_id"`
	Date           time.Time          `json:"date"`
	Material       string             `json:"material"`
	Qty            float64            `json:"qty"`
	Unit           string             `json:"unit"`
	DocumentNumber string             `json:"document_number"`
	Source         string             `json:"source"` // "MANUAL" | "CV"
	CVConfidence   float64            `json:"cv_confidence"`
	Documents      []MaterialDocument `json:"documents,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}
