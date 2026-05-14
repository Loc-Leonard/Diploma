package inspector

import (
	"time"

	"github.com/Loc-Leonard/Diploma/backend/internal/models"
)

// InspectorDashboardCheck - DTO для проверок на дашборде
type InspectorDashboardCheck struct {
	ID         uint      `json:"id"`
	ObjectID   uint      `json:"object_id"`
	ObjectName string    `json:"object_name"`
	City       string    `json:"city"`
	Address    string    `json:"address"`
	Status     string    `json:"status"`
	PlannedAt  time.Time `json:"planned_at"`
	IssuesOpen int       `json:"issues_open"`
}

// InspectorDashboardObject - DTO для объектов на дашборде
type InspectorDashboardObject struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	City          string `json:"city"`
	Address       string `json:"address"`
	ForemanName   string `json:"foreman_name"`
	ActiveChecks  int    `json:"active_checks"`
	OverdueChecks int    `json:"overdue_checks"`
	OpenIssues    int    `json:"open_issues"`
}

// InspectorObjectListItem - DTO для списка объектов
type InspectorObjectListItem struct {
	ID               uint                `json:"id"`
	Name             string              `json:"name"`
	City             string              `json:"city"`
	Address          string              `json:"address"`
	Status           models.ObjectStatus `json:"status"`
	ForemanName      string              `json:"foreman_name"`
	PlannedStartDate *time.Time          `json:"planned_start_date"`
	HasPendingAction bool                `json:"has_pending_action"`
	Lat              float64             `json:"lat"`
	Lng              float64             `json:"lng"`
	Progress         float64             `json:"progress"`
}

// PersonDTO - DTO персоны
type PersonDTO struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
}

// ObjectDTO - DTO объекта
type ObjectDTO struct {
	ID                     uint                `json:"id"`
	Name                   string              `json:"name"`
	City                   string              `json:"city"`
	Address                string              `json:"address"`
	Description            string              `json:"description"`
	Status                 models.ObjectStatus `json:"status"`
	Lat                    float64             `json:"lat"`
	Lng                    float64             `json:"lng"`
	PlannedStartDate       time.Time           `json:"planned_start_date"`
	PlannedEndDate         time.Time           `json:"planned_end_date"`
	ActualStartDate        *time.Time          `json:"actual_start_date"`
	Customer               *PersonDTO          `json:"customer,omitempty"`
	Foreman                *PersonDTO          `json:"foreman,omitempty"`
	Inspector              *PersonDTO          `json:"inspector,omitempty"`
	InitChecklistJSON      string              `json:"init_checklist_json"`
	InitActFilePath        string              `json:"init_act_file_path"`
	ActivationRejectReason string              `json:"activation_reject_reason"`
}

// ObjectDetailResponse - ответ с деталями объекта
type ObjectDetailResponse struct {
	Object     ObjectDTO                 `json:"object"`
	WorkItems  []models.WorkItem         `json:"work_items"`
	Deliveries []models.MaterialDelivery `json:"deliveries"`
}

// ActivationDecisionRequest - запрос на решение об активации
type ActivationDecisionRequest struct {
	Decision        string `json:"decision" binding:"required"`
	RejectionReason string `json:"rejection_reason"`
}

// DocumentDTO - DTO для документа
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

// DocumentUploadResponse - ответ после загрузки документа
type DocumentUploadResponse struct {
	Status     string `json:"status"`
	DocumentID uint   `json:"document_id"`
	FileName   string `json:"file_name"`
	FilePath   string `json:"file_path"`
	CVStatus   string `json:"cv_status"`
}
