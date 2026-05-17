package models

import "time"

type MaterialDocumentType string

const (
	MaterialDocumentTypeTTN             MaterialDocumentType = "TTN"
	MaterialDocumentTypeQualityPassport MaterialDocumentType = "QUALITY_PASSPORT"
	MaterialDocumentTypePhoto           MaterialDocumentType = "PHOTO"
	MaterialDocumentTypeOther           MaterialDocumentType = "OTHER"
)

type CVProcessingStatus string

const (
	CVProcessingStatusPending CVProcessingStatus = "PENDING"
	CVProcessingStatusDone    CVProcessingStatus = "DONE"
	CVProcessingStatusFailed  CVProcessingStatus = "FAILED"
)

type MaterialDocument struct {
	ID               uint                 `gorm:"primaryKey" json:"id"`
	ObjectID         *uint                `json:"object_id"`
	DeliveryID       *uint                `json:"delivery_id"`
	IssueID          *uint                `gorm:"index" json:"issue_id,omitempty"`
	UploadedBy       *uint                `json:"uploaded_by,omitempty"`
	DocumentType     MaterialDocumentType `json:"document_type"`
	StoragePath      string               `json:"storage_path"`
	OriginalFileName string               `json:"original_file_name"`
	MimeType         string               `json:"mime_type"`
	CVStatus         CVProcessingStatus   `json:"cv_status"`
	CVConfidence     float64              `json:"cv_confidence"`
	CVPayloadJSON    string               `json:"cv_payload_json"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}
