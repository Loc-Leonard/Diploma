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
	ID               uint                `json:"id"`
	Name             string              `json:"name"`
	City             string              `json:"city"`
	Address          string              `json:"address"`
	Description      string              `json:"description"`
	Status           models.ObjectStatus `json:"status"`
	Lat              float64             `json:"lat"`
	Lng              float64             `json:"lng"`
	PlannedStartDate *time.Time          `json:"planned_start_date,omitempty"`
	PlannedEndDate   *time.Time          `json:"planned_end_date,omitempty"`
	Customer         *ObjectPersonDTO    `json:"customer,omitempty"`
	Foreman          *ObjectPersonDTO    `json:"foreman,omitempty"`
	Inspector        *ObjectPersonDTO    `json:"inspector,omitempty"`
}
