package foreman

import (
	"time"

	"github.com/Loc-Leonard/Diploma/backend/internal/models"
)

type ForemanObjectDTO struct {
	ID               uint                `json:"id"`
	Name             string              `json:"name"`
	City             string              `json:"city"`
	Address          string              `json:"address"`
	Status           models.ObjectStatus `json:"status"`
	PlannedStartDate time.Time           `json:"planned_start_date"`
	PlannedEndDate   time.Time           `json:"planned_end_date"`
	Lng              float64             `json:"lng"`
	Lat              float64             `json:"lat"`
	Progress         float64             `json:"progress"`
}

type WorkItemDTO struct {
	ID               uint       `json:"id"`
	ObjectID         uint       `json:"object_id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Unit             string     `json:"unit"`
	PlanQty          float64    `json:"plan_qty"`
	PlannedStartDate *time.Time `json:"planned_start_date,omitempty"`
	PlannedEndDate   *time.Time `json:"planned_end_date,omitempty"`
	SortOrder        int        `json:"sort_order"`
	Status           string     `json:"status"`
}

type DeliveryDTO struct {
	ID             uint      `json:"id"`
	ObjectID       uint      `json:"object_id"`
	ForemanID      uint      `json:"foreman_id"`
	Date           time.Time `json:"date"`
	Material       string    `json:"material"`
	Qty            float64   `json:"qty"`
	Unit           string    `json:"unit"`
	DocumentNumber string    `json:"document_number"`
	Source         string    `json:"source"`
	CVConfidence   float64   `json:"cv_confidence"`
}

type DocumentDTO struct {
	ID               uint      `json:"id"`
	DocumentType     string    `json:"document_type"`
	OriginalFileName string    `json:"original_file_name"`
	MimeType         string    `json:"mime_type"`
	CVStatus         string    `json:"cv_status"`
	CVConfidence     float64   `json:"cv_confidence,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UploadedBy       string    `json:"uploaded_by"`
	DownloadURL      string    `json:"download_url"`
}

type DocumentUploadResponse struct {
	Status     string `json:"status"`
	DocumentID uint   `json:"document_id"`
	FileName   string `json:"file_name"`
	FilePath   string `json:"file_path"`
	CVStatus   string `json:"cv_status"`
}

type ObjectDetailResponse struct {
	Object     ObjectDTO     `json:"object"`
	WorkItems  []WorkItemDTO `json:"work_items"`
	Deliveries []DeliveryDTO `json:"deliveries"`
}

type ObjectDTO struct {
	ID               uint                `json:"id"`
	Name             string              `json:"name"`
	City             string              `json:"city"`
	Address          string              `json:"address"`
	Description      string              `json:"description"`
	Status           models.ObjectStatus `json:"status"`
	Lat              float64             `json:"lat"`
	Lng              float64             `json:"lng"`
	PlannedStartDate time.Time           `json:"planned_start_date"`
	PlannedEndDate   time.Time           `json:"planned_end_date"`
	ActualStartDate  *time.Time          `json:"actual_start_date"`
	Customer         *PersonDTO          `json:"customer,omitempty"`
	Foreman          *PersonDTO          `json:"foreman,omitempty"`
	Inspector        *PersonDTO          `json:"inspector,omitempty"`
}

type PersonDTO struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
}
