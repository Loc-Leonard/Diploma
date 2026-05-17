package models

import "time"

type IssueType string

const (
	IssueTypeRemark    IssueType = "remark"
	IssueTypeViolation IssueType = "violation"
)

type IssueStatus string

const (
	IssueStatusOpen              IssueStatus = "open"
	IssueStatusInProgress        IssueStatus = "in_progress"
	IssueStatusResolvedByForeman IssueStatus = "resolved_by_foreman"
	IssueStatusAccepted          IssueStatus = "accepted"
	IssueStatusRejected          IssueStatus = "rejected"
)

type Issue struct {
	ID                uint        `gorm:"primaryKey" json:"id"`
	ObjectID          uint        `gorm:"index;not null" json:"object_id"`
	Object            Object      `json:"-"`
	Type              IssueType   `gorm:"type:varchar(20);index;not null" json:"type"`
	Status            IssueStatus `gorm:"type:varchar(32);index;not null" json:"status"`
	Title             string      `json:"title"`
	Description       string      `gorm:"type:text;not null" json:"description"`
	DueDate           *time.Time  `gorm:"index" json:"due_date"`
	AuthorID          uint        `gorm:"index;not null" json:"author_id"`
	Author            User        `json:"-"`
	AuthorRole        Role        `gorm:"type:varchar(20);not null" json:"author_role"`
	ClassifierCode    *string     `gorm:"type:varchar(128)" json:"classifier_code,omitempty"`
	Comment           string      `gorm:"type:text" json:"comment"`
	ResolvedByID      *uint       `gorm:"index" json:"resolved_by_id,omitempty"`
	ResolvedAt        *time.Time  `json:"resolved_at,omitempty"`
	ResolutionComment string      `gorm:"type:text" json:"resolution_comment"`

	ReviewedByID    *uint      `gorm:"index" json:"reviewed_by_id.omitempty"`
	AcceptedAt      *time.Time `json:"accepted_at,omitempty"`
	RejectionReason string     `gorm:"type:text" json:"rejection_reason"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Comments      []IssueComment       `gorm:"foreignKey:IssueID" json:"-"`
	StatusHistory []IssueStatusHistory `gorm:"foreignKey:IssueID" json:"-"`
	Attachments   []MaterialDocument   `gorm:"foreignKey:IssueID" json:"-"`
}

type IssueCommentType string

const (
	IssueCommentTypeGeneral    IssueCommentType = "general"
	IssueCommentTypeResolution IssueCommentType = "resolution"
	IssueCommentTypeReview     IssueCommentType = "review"
)

type IssueComment struct {
	ID uint `gorm:"primaryKey" json:"id"`

	IssueID uint  `gorm:"index;not null" json:"issue_id"`
	Issue   Issue `json:"-"`

	AuthorID   uint `gorm:"index;not null" json:"author_id"`
	Author     User `json:"-"`
	AuthorRole Role `gorm:"type:varchar(20);not null" json:"author_role"`

	CommentType IssueCommentType `gorm:"type:varchar(20);not null" json:"comment_type"`
	Comment     string           `gorm:"type:text;not null" json:"comment"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type IssueStatusHistory struct {
	ID uint `gorm:"primaryKey" json:"id"`

	IssueID uint  `gorm:"index;not null" json:"issue_id"`
	Issue   Issue `json:"-"`

	FromStatus *IssueStatus `gorm:"type:varchar(32)" json:"from_status,omitempty"`
	ToStatus   IssueStatus  `gorm:"type:varchar(32);not null" json:"to_status"`

	ChangedByID   uint `gorm:"index;not null" json:"changed_by_id"`
	ChangedBy     User `json:"-"`
	ChangedByRole Role `gorm:"type:varchar(20);not null" json:"changed_by_role"`

	Comment string `gorm:"type:text" json:"comment"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
