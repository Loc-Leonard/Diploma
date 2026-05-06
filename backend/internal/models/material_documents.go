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
	DeliveryID       uint                 `json:"delivery_id"`
	DocumentType     MaterialDocumentType `json:"document_type"`
	StoragePath      string               `json:"storage_path"`
	OriginalFileName string               `json:"original_file_name"`
	MimeType         string               `json:"mime_type"`
	CVStatus         CVProcessingStatus   `json:"cv_status"`
	CVPayloadJSON    string               `json:"cv_payload_json"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}
