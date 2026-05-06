package objectcore

import (
	"time"

	"github.com/Loc-Leonard/Diploma/internal/models"
)

type ObjectPersonDTO struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
}

type ObjectCoreDTO struct {
	ID                     uint                `json:"id"`
	Name                   string              `json:"name"`
	City                   string              `json:"city"`
	Address                string              `json:"address"`
	Description            string              `json:"description"`
	Status                 models.ObjectStatus `json:"status"`
	Lat                    float64             `json:"lat"`
	Lng                    float64             `json:"lng"`
	PlannedStartDate       *time.Time          `json:"planned_start_date,omitempty"`
	PlannedEndDate         *time.Time          `json:"planned_end_date,omitempty"`
	Customer               *ObjectPersonDTO    `json:"customer,omitempty"`
	Foreman                *ObjectPersonDTO    `json:"foreman,omitempty"`
	Inspector              *ObjectPersonDTO    `json:"inspector,omitempty"`
	InitActFilePath        string              `json:"init_act_file_path"`
	InitChecklistJSON      string              `json:"init_checklist_json"`
	ActualStartDate        *time.Time          `json:"actual_start_date,omitempty"`
	ActivationRejectReason string              `json:"activation_reject_reason"`
	Progress               float64             `json:"progress"`
}

type ObjectDetailDTO struct {
	Object     ObjectCoreDTO             `json:"object"`
	WorkItems  []models.WorkItem         `json:"work_items"`
	Deliveries []models.MaterialDelivery `json:"deliveries"`
}
