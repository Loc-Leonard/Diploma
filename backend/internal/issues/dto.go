package issues

import (
	"time"

	"github.com/Loc-Leonard/Diploma/backend/internal/models"
)

type CreateIssueRequest struct {
	Title          string     `json:"title"`
	Description    string     `json:"description" binding:"required"`
	DueDate        *time.Time `json:"due_date" binding:"required"`
	ClassifierCode *string    `json:"classifier_code"`
	Comment        string     `json:"comment"`
}

type UpdateIssueRequest struct {
	Status *models.IssueStatus `json:"status"`
}

type ResolveIssueRequest struct {
	Comment string `json:"comment" binding:"required"`
}

type ReviewIssueRequest struct {
	Decision string `json:"decision" binding:"required"`
	Comment  string `json:"comment"`
}

type AddIssueCommentRequest struct {
	Comment string `json:"comment" binding:"required"`
}

type IssueAttachmentUploadResponse struct {
	Status     string `json:"status"`
	DocumentID uint   `json:"document_id"`
	FileName   string `json:"file_name"`
	FilePath   string `json:"file_path"`
}

type IssuePersonDTO struct {
	ID       uint        `json:"id"`
	FullName string      `json:"full_name"`
	Role     models.Role `json:"role"`
}

type IssueAttachmentDTO struct {
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

type IssueCommentDTO struct {
	ID          uint      `json:"id"`
	AuthorID    uint      `json:"author_id"`
	AuthorName  string    `json:"author_name"`
	AuthorRole  string    `json:"author_role"`
	Comment     string    `json:"comment"`
	CommentType string    `json:"comment_type"`
	CreatedAt   time.Time `json:"created_at"`
}

type IssueStatusHistoryDTO struct {
	ID            uint      `json:"id"`
	FromStatus    *string   `json:"from_status,omitempty"`
	ToStatus      string    `json:"to_status"`
	ChangedByID   uint      `json:"changed_by_id"`
	ChangedByName string    `json:"changed_by_name"`
	ChangedByRole string    `json:"changed_by_role"`
	Comment       string    `json:"comment"`
	CreatedAt     time.Time `json:"created_at"`
}

type IssueListItemDTO struct {
	ID              uint                 `json:"id"`
	ObjectID        uint                 `json:"object_id"`
	Type            models.IssueType     `json:"type"`
	Status          models.IssueStatus   `json:"status"`
	DisplayStatus   string               `json:"display_status"`
	Title           string               `json:"title"`
	Description     string               `json:"description"`
	DueDate         *time.Time           `json:"due_date"`
	AuthorID        uint                 `json:"author_id"`
	AuthorName      string               `json:"author_name"`
	AuthorRole      models.Role          `json:"author_role"`
	ClassifierCode  *string              `json:"classifier_code,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	ResolvedAt      *time.Time           `json:"resolved_at,omitempty"`
	AcceptedAt      *time.Time           `json:"accepted_at,omitempty"`
	RejectionReason string               `json:"rejection_reason,omitempty"`
	IsOverdue       bool                 `json:"is_overdue"`
	Attachments     []IssueAttachmentDTO `json:"attachments"`
}

type IssueDetailDTO struct {
	IssueListItemDTO
	Comment           string                  `json:"comment"`
	ResolutionComment string                  `json:"resolution_comment"`
	Comments          []IssueCommentDTO       `json:"comments"`
	StatusHistory     []IssueStatusHistoryDTO `json:"status_history"`
}
