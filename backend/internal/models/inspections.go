package models

import "time"

type InspectionStatus string

const (
	InspectionStatusPlanned    InspectionStatus = "PLANNED"
	InspectionStatusInProgress InspectionStatus = "IN_PROGRESS"
	InspectionStatusFinished   InspectionStatus = "FINISHED"
	InspectionStatusOverdue    InspectionStatus = "OVERDUE"
)

type Inspection struct {
	ID          uint             `gorm:"primaryKey" json:"id"`
	ObjectID    uint             `json:"object_id"`
	Object      Object           `json:"object"` // для Preload
	InspectorID uint             `json:"inspector_id"`
	Inspector   User             `json:"inspector"` // опционально, тоже через Preload
	Status      InspectionStatus `json:"status"`
	PlannedAt   time.Time        `json:"planned_at"`
	FinishedAt  *time.Time       `json:"finished_at,omitempty"`
	IssuesOpen  int              `json:"issues_open"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
