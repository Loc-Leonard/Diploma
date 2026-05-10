package customer

import (
	"time"

	"github.com/Loc-Leonard/Diploma/backend/internal/models"
)

// DashboardObjectDTO - DTO объекта для дашборда
type DashboardObjectDTO struct {
	ID       uint                `json:"id"`
	Name     string              `json:"name"`
	City     string              `json:"city"`
	Address  string              `json:"address"`
	Status   models.ObjectStatus `json:"status"`
	Progress float64             `json:"progress"`
	Foreman  *struct {
		ID       uint   `json:"id"`
		FullName string `json:"full_name"`
	} `json:"foreman,omitempty"`
	PlannedStartDate       *time.Time `json:"planned_start_date"`
	PlannedEndDate         *time.Time `json:"planned_end_date"`
	Lat                    float64    `json:"lat"`
	Lng                    float64    `json:"lng"`
	ActivationRejectReason string     `json:"activation_reject_reason"`
}

// DashboardForemanDTO - DTO прораба для дашборда
type DashboardForemanDTO struct {
	ID            uint   `json:"id"`
	FullName      string `json:"full_name"`
	City          string `json:"city"`
	CurrentObject *struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	} `json:"current_object,omitempty"`
}

// CreateObjectRequest - DTO для создания объекта
type CreateObjectRequest struct {
	Name             string     `json:"name" binding:"required"`
	Address          string     `json:"address" binding:"required"`
	City             string     `json:"city" binding:"required"`
	Description      string     `json:"description"`
	Lat              float64    `json:"lat"`
	Lng              float64    `json:"lng"`
	PlannedStartDate *time.Time `json:"planned_start_date"`
	PlannedEndDate   *time.Time `json:"planned_end_date"`
	ForemanUserID    uint       `json:"foreman_user_id" binding:"required"`
	InspectorUserID  uint       `json:"inspector_user_id" binding:"required"`
}

// ActivateObjectRequest - DTO для активации объекта
type ActivateObjectRequest struct {
	ChecklistJSON string  `json:"checklist_json" binding:"required"`
	ActFilePath   *string `json:"act_file_path"`
}

// WorkItemInput - DTO для работы с задачами
type WorkItemInput struct {
	Name             string     `json:"name" binding:"required"`
	Description      string     `json:"description"`
	Unit             string     `json:"unit"`
	PlanQty          float64    `json:"plan_qty"`
	PlannedStartDate *time.Time `json:"planned_start_date"`
	PlannedEndDate   *time.Time `json:"planned_end_date"`
	SortOrder        int        `json:"sort_order"`
	DependsOnID      *uint      `json:"depends_on_id"`
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

// UploadDocumentRequest - запрос на загрузку документа
type UploadDocumentRequest struct {
	DocumentType string `json:"document_type" binding:"required"`
	Description  string `json:"description"`
}

// DocumentUploadResponse - ответ после загрузки документа
type DocumentUploadResponse struct {
	Status     string `json:"status"`
	DocumentID uint   `json:"document_id"`
	FileName   string `json:"file_name"`
	FilePath   string `json:"file_path"`
	CVStatus   string `json:"cv_status"`
}
