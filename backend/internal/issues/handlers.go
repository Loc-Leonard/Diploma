package issues

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/backend/internal/auth"
	"github.com/Loc-Leonard/Diploma/backend/internal/models"
)

type Handler struct {
	db          *gorm.DB
	storageRoot string
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB, storageRoot string) {
	h := &Handler{db: db, storageRoot: storageRoot}

	customer := r.Group("/customer")
	customer.Use(auth.AuthMiddleware(), auth.MustChangePasswordMiddleware(db), auth.CustomerOnly())
	{
		customer.GET("/objects/:id/issues", h.ListCustomerObjectIssues)
		customer.POST("/objects/:id/issues", h.CreateRemark)
		customer.POST("/issues/:id/review", h.ReviewRemark)
	}

	foreman := r.Group("/foreman")
	foreman.Use(auth.AuthMiddleware(), auth.MustChangePasswordMiddleware(db), auth.ForemanOnly())
	{
		foreman.GET("/objects/:id/issues", h.ListForemanObjectIssues)
		foreman.POST("/issues/:id/resolve", h.ResolveIssue)
		foreman.PATCH("/issues/:id", h.UpdateForemanIssue)
	}

	inspector := r.Group("/inspector")
	inspector.Use(auth.AuthMiddleware(), auth.MustChangePasswordMiddleware(db), auth.InspectorOnly())
	{
		inspector.GET("/objects/:id/issues", h.ListInspectorObjectIssues)
		inspector.POST("/objects/:id/issues", h.CreateViolation)
		inspector.POST("/issues/:id/review", h.ReviewViolation)
	}

	common := r.Group("/issues")
	common.Use(auth.AuthMiddleware(), auth.MustChangePasswordMiddleware(db))
	{
		common.GET("/:id", h.GetIssue)
		common.GET("/:id/history", h.GetIssueHistory)
		common.POST("/:id/comments", h.AddComment)
		common.POST("/:id/attachments", h.UploadAttachment)
	}
}

func (h *Handler) ListCustomerObjectIssues(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	var obj models.Object
	if err := h.db.Where("id = ? AND customer_control_user_id = ?", objectID, userID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	h.listByObject(c, obj.ID)
}

func (h *Handler) ListForemanObjectIssues(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	var obj models.Object
	if err := h.db.Where("id = ? AND foreman_user_id = ?", objectID, userID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	h.listByObject(c, obj.ID)
}

func (h *Handler) ListInspectorObjectIssues(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	var obj models.Object
	if err := h.db.Where("id = ? AND inspector_user_id = ?", objectID, userID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	h.listByObject(c, obj.ID)
}

func (h *Handler) listByObject(c *gin.Context, objectID uint) {
	var issues []models.Issue
	if err := h.db.
		Preload("Attachments").
		Where("object_id = ?", objectID).
		Order("created_at DESC").
		Find(&issues).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	c.JSON(http.StatusOK, h.toIssueListDTOs(issues))
}

func (h *Handler) CreateRemark(c *gin.Context) {
	h.createIssueByRole(c, models.IssueTypeRemark, models.RoleCustomer)
}

func (h *Handler) CreateViolation(c *gin.Context) {
	h.createIssueByRole(c, models.IssueTypeViolation, models.RoleInspector)
}

func (h *Handler) createIssueByRole(c *gin.Context, issueType models.IssueType, role models.Role) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	var req CreateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if strings.TrimSpace(req.Description) == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "description is required"})
		return
	}
	if req.DueDate == nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "due_date is required"})
		return
	}
	if issueType == models.IssueTypeViolation && (req.ClassifierCode == nil || strings.TrimSpace(*req.ClassifierCode) == "") {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "classifier_code is required"})
		return
	}

	var obj models.Object
	query := h.db.Where("id = ?", objectID)
	switch role {
	case models.RoleCustomer:
		query = query.Where("customer_control_user_id = ?", userID)
	case models.RoleInspector:
		query = query.Where("inspector_user_id = ?", userID)
	}
	if err := query.First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	issue := models.Issue{
		ObjectID:       obj.ID,
		Type:           issueType,
		Status:         models.IssueStatusOpen,
		Title:          strings.TrimSpace(req.Title),
		Description:    strings.TrimSpace(req.Description),
		DueDate:        req.DueDate,
		AuthorID:       userID,
		AuthorRole:     role,
		ClassifierCode: req.ClassifierCode,
		Comment:        strings.TrimSpace(req.Comment),
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&issue).Error; err != nil {
			return err
		}

		history := models.IssueStatusHistory{
			IssueID:       issue.ID,
			FromStatus:    nil,
			ToStatus:      models.IssueStatusOpen,
			ChangedByID:   userID,
			ChangedByRole: role,
			Comment:       "issue created",
		}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}

		if strings.TrimSpace(req.Comment) != "" {
			comment := models.IssueComment{
				IssueID:     issue.ID,
				AuthorID:    userID,
				AuthorRole:  role,
				CommentType: models.IssueCommentTypeGeneral,
				Comment:     strings.TrimSpace(req.Comment),
			}
			if err := tx.Create(&comment).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	var created models.Issue
	if err := h.db.Preload("Attachments").First(&created, issue.ID).Error; err != nil {
		c.JSON(http.StatusCreated, gin.H{"id": issue.ID})
		return
	}

	c.JSON(http.StatusCreated, h.toIssueListDTO(created))
}

func (h *Handler) UpdateForemanIssue(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	issue, obj, ok := h.loadIssueForForeman(c, userID)
	if !ok || obj == nil {
		return
	}

	var req UpdateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}
	if req.Status == nil || *req.Status != models.IssueStatusInProgress {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "only in_progress status is allowed"})
		return
	}
	if issue.Status != models.IssueStatusOpen && issue.Status != models.IssueStatusRejected {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "issue cannot be moved to in_progress"})
		return
	}

	from := issue.Status
	issue.Status = models.IssueStatusInProgress

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&issue).Error; err != nil {
			return err
		}
		history := models.IssueStatusHistory{
			IssueID:       issue.ID,
			FromStatus:    &from,
			ToStatus:      issue.Status,
			ChangedByID:   userID,
			ChangedByRole: models.RoleForeman,
			Comment:       "work started by foreman",
		}
		return tx.Create(&history).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Status: "ok"})
}

func (h *Handler) ResolveIssue(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	issue, _, ok := h.loadIssueForForeman(c, userID)
	if !ok {
		return
	}

	var req ResolveIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}
	if strings.TrimSpace(req.Comment) == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "comment is required"})
		return
	}
	if !(issue.Status == models.IssueStatusOpen || issue.Status == models.IssueStatusInProgress || issue.Status == models.IssueStatusRejected) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "issue cannot be resolved"})
		return
	}

	now := time.Now()
	from := issue.Status
	issue.Status = models.IssueStatusResolvedByForeman
	issue.ResolvedByID = &userID
	issue.ResolvedAt = &now
	issue.ResolutionComment = strings.TrimSpace(req.Comment)

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&issue).Error; err != nil {
			return err
		}

		comment := models.IssueComment{
			IssueID:     issue.ID,
			AuthorID:    userID,
			AuthorRole:  models.RoleForeman,
			CommentType: models.IssueCommentTypeResolution,
			Comment:     strings.TrimSpace(req.Comment),
		}
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}

		history := models.IssueStatusHistory{
			IssueID:       issue.ID,
			FromStatus:    &from,
			ToStatus:      issue.Status,
			ChangedByID:   userID,
			ChangedByRole: models.RoleForeman,
			Comment:       strings.TrimSpace(req.Comment),
		}
		return tx.Create(&history).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Status: "ok"})
}

func (h *Handler) ReviewRemark(c *gin.Context) {
	h.reviewIssue(c, models.RoleCustomer, models.IssueTypeRemark)
}

func (h *Handler) ReviewViolation(c *gin.Context) {
	h.reviewIssue(c, models.RoleInspector, models.IssueTypeViolation)
}

func (h *Handler) reviewIssue(c *gin.Context, reviewerRole models.Role, issueType models.IssueType) {
	userID := auth.UserIDFromContext(c)
	issueID := c.Param("id")

	var issue models.Issue
	if err := h.db.First(&issue, issueID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "issue not found"})
		return
	}
	if issue.Type != issueType {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "issue type mismatch"})
		return
	}
	if issue.Status != models.IssueStatusResolvedByForeman {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "issue is not ready for review"})
		return
	}

	var obj models.Object
	query := h.db.Where("id = ?", issue.ObjectID)
	switch reviewerRole {
	case models.RoleCustomer:
		query = query.Where("customer_control_user_id = ?", userID)
	case models.RoleInspector:
		query = query.Where("inspector_user_id = ?", userID)
	}
	if err := query.First(&obj).Error; err != nil {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "access denied"})
		return
	}

	var req ReviewIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	decision := strings.ToUpper(strings.TrimSpace(req.Decision))
	now := time.Now()
	from := issue.Status

	switch decision {
	case "ACCEPT":
		issue.Status = models.IssueStatusAccepted
		issue.AcceptedAt = &now
		issue.ReviewedByID = &userID
	case "REJECT":
		if strings.TrimSpace(req.Comment) == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "comment is required for rejection"})
			return
		}
		issue.Status = models.IssueStatusRejected
		issue.ReviewedByID = &userID
		issue.RejectionReason = strings.TrimSpace(req.Comment)
	default:
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid decision"})
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&issue).Error; err != nil {
			return err
		}

		if strings.TrimSpace(req.Comment) != "" {
			comment := models.IssueComment{
				IssueID:     issue.ID,
				AuthorID:    userID,
				AuthorRole:  reviewerRole,
				CommentType: models.IssueCommentTypeReview,
				Comment:     strings.TrimSpace(req.Comment),
			}
			if err := tx.Create(&comment).Error; err != nil {
				return err
			}
		}

		history := models.IssueStatusHistory{
			IssueID:       issue.ID,
			FromStatus:    &from,
			ToStatus:      issue.Status,
			ChangedByID:   userID,
			ChangedByRole: reviewerRole,
			Comment:       strings.TrimSpace(req.Comment),
		}
		return tx.Create(&history).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Status: "ok"})
}

func (h *Handler) GetIssue(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	issueID := c.Param("id")

	issue, ok := h.loadIssueWithAccess(c, issueID, userID)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, h.toIssueDetailDTO(issue))
}

func (h *Handler) GetIssueHistory(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	issueID := c.Param("id")

	issue, ok := h.loadIssueWithAccess(c, issueID, userID)
	if !ok {
		return
	}

	out := make([]IssueStatusHistoryDTO, 0, len(issue.StatusHistory))
	users := h.loadUserNamesFromHistory(issue.StatusHistory)

	for _, item := range issue.StatusHistory {
		var from *string
		if item.FromStatus != nil {
			v := string(*item.FromStatus)
			from = &v
		}
		out = append(out, IssueStatusHistoryDTO{
			ID:            item.ID,
			FromStatus:    from,
			ToStatus:      string(item.ToStatus),
			ChangedByID:   item.ChangedByID,
			ChangedByName: users[item.ChangedByID],
			ChangedByRole: string(item.ChangedByRole),
			Comment:       item.Comment,
			CreatedAt:     item.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, out)
}

func (h *Handler) AddComment(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	role := models.Role(c.GetString("role"))
	issueID := c.Param("id")

	issue, ok := h.loadIssueWithAccess(c, issueID, userID)
	if !ok {
		return
	}

	var req AddIssueCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}
	if strings.TrimSpace(req.Comment) == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "comment is required"})
		return
	}

	comment := models.IssueComment{
		IssueID:     issue.ID,
		AuthorID:    userID,
		AuthorRole:  role,
		CommentType: models.IssueCommentTypeGeneral,
		Comment:     strings.TrimSpace(req.Comment),
	}

	if err := h.db.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{Status: "ok"})
}

func (h *Handler) UploadAttachment(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	issueID := c.Param("id")

	issue, ok := h.loadIssueWithAccess(c, issueID, userID)
	if !ok {
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "no file provided"})
		return
	}

	const maxFileSize = 10 * 1024 * 1024
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "file too large (max 10MB)"})
		return
	}

	allowedMimeTypes := []string{
		"image/jpeg", "image/png", "image/jpg", "image/gif",
		"application/pdf",
		"application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}

	fh, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "cannot read file"})
		return
	}
	defer fh.Close()

	buffer := make([]byte, 512)
	_, err = fh.Read(buffer)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "cannot read file"})
		return
	}
	fileType := http.DetectContentType(buffer)
	if !slices.Contains(allowedMimeTypes, fileType) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "file type not allowed"})
		return
	}

	objectDir := filepath.Join(h.storageRoot, "issues", fmt.Sprintf("%d", issue.ID), "attachments")
	if err := os.MkdirAll(objectDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "cannot create directory"})
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	originalFilename := filepath.Base(file.Filename)
	cleanFilename := strings.ReplaceAll(originalFilename, " ", "_")
	filename := timestamp + "_" + cleanFilename
	filePath := filepath.Join(objectDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "cannot save file"})
		return
	}

	docTypeStr := c.PostForm("document_type")
	if docTypeStr == "" {
		docTypeStr = "OTHER"
	}
	objectID := issue.ObjectID
	doc := models.MaterialDocument{
		ObjectID:         &objectID,
		IssueID:          &issue.ID,
		UploadedBy:       &userID,
		DocumentType:     models.MaterialDocumentType(docTypeStr),
		StoragePath:      filePath,
		OriginalFileName: originalFilename,
		MimeType:         fileType,
		CVStatus:         models.CVProcessingStatusPending,
		CVPayloadJSON:    "",
		CVConfidence:     0,
	}

	if err := h.db.Create(&doc).Error; err != nil {
		_ = os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	c.JSON(http.StatusCreated, IssueAttachmentUploadResponse{
		Status:     "uploaded",
		DocumentID: doc.ID,
		FileName:   originalFilename,
		FilePath:   filePath,
	})
}

func (h *Handler) loadIssueForForeman(c *gin.Context, userID uint) (models.Issue, *models.Object, bool) {
	issueID := c.Param("id")
	var issue models.Issue
	if err := h.db.First(&issue, issueID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "issue not found"})
		return issue, nil, false
	}
	var obj models.Object
	if err := h.db.Where("id = ? AND foreman_user_id = ?", issue.ObjectID, userID).First(&obj).Error; err != nil {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "access denied"})
		return issue, nil, false
	}
	return issue, &obj, true
}

func (h *Handler) loadIssueWithAccess(c *gin.Context, issueID string, userID uint) (models.Issue, bool) {
	var issue models.Issue
	if err := h.db.
		Preload("Attachments").
		Preload("Comments").
		Preload("StatusHistory", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC") }).
		First(&issue, issueID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "issue not found"})
		return issue, false
	}

	var obj models.Object
	if err := h.db.First(&obj, issue.ObjectID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return issue, false
	}

	if obj.CustomerControlUserID != userID && obj.ForemanUserID != userID && obj.InspectorUserID != userID {
		var user models.User
		if err := h.db.First(&user, userID).Error; err != nil || user.Role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "access denied"})
			return issue, false
		}
	}

	return issue, true
}

func (h *Handler) toIssueListDTOs(issues []models.Issue) []IssueListItemDTO {
	out := make([]IssueListItemDTO, 0, len(issues))
	for _, issue := range issues {
		out = append(out, h.toIssueListDTO(issue))
	}
	return out
}

func (h *Handler) toIssueListDTO(issue models.Issue) IssueListItemDTO {
	authorName := h.getUserName(issue.AuthorID)
	return IssueListItemDTO{
		ID:              issue.ID,
		ObjectID:        issue.ObjectID,
		Type:            issue.Type,
		Status:          issue.Status,
		DisplayStatus:   h.displayStatus(issue),
		Title:           issue.Title,
		Description:     issue.Description,
		DueDate:         issue.DueDate,
		AuthorID:        issue.AuthorID,
		AuthorName:      authorName,
		AuthorRole:      issue.AuthorRole,
		ClassifierCode:  issue.ClassifierCode,
		CreatedAt:       issue.CreatedAt,
		UpdatedAt:       issue.UpdatedAt,
		ResolvedAt:      issue.ResolvedAt,
		AcceptedAt:      issue.AcceptedAt,
		RejectionReason: issue.RejectionReason,
		IsOverdue:       h.isOverdue(issue),
		Attachments:     h.toAttachmentDTOs(issue.Attachments),
	}
}

func (h *Handler) toIssueDetailDTO(issue models.Issue) IssueDetailDTO {
	base := h.toIssueListDTO(issue)
	usersMap := h.loadCommentUserNames(issue.Comments)
	historyUsersMap := h.loadUserNamesFromHistory(issue.StatusHistory)

	comments := make([]IssueCommentDTO, 0, len(issue.Comments))
	for _, item := range issue.Comments {
		comments = append(comments, IssueCommentDTO{
			ID:          item.ID,
			AuthorID:    item.AuthorID,
			AuthorName:  usersMap[item.AuthorID],
			AuthorRole:  string(item.AuthorRole),
			Comment:     item.Comment,
			CommentType: string(item.CommentType),
			CreatedAt:   item.CreatedAt,
		})
	}

	history := make([]IssueStatusHistoryDTO, 0, len(issue.StatusHistory))
	for _, item := range issue.StatusHistory {
		var from *string
		if item.FromStatus != nil {
			v := string(*item.FromStatus)
			from = &v
		}
		history = append(history, IssueStatusHistoryDTO{
			ID:            item.ID,
			FromStatus:    from,
			ToStatus:      string(item.ToStatus),
			ChangedByID:   item.ChangedByID,
			ChangedByName: historyUsersMap[item.ChangedByID],
			ChangedByRole: string(item.ChangedByRole),
			Comment:       item.Comment,
			CreatedAt:     item.CreatedAt,
		})
	}

	return IssueDetailDTO{
		IssueListItemDTO:  base,
		Comment:           issue.Comment,
		ResolutionComment: issue.ResolutionComment,
		Comments:          comments,
		StatusHistory:     history,
	}
}

func (h *Handler) toAttachmentDTOs(items []models.MaterialDocument) []IssueAttachmentDTO {
	out := make([]IssueAttachmentDTO, 0, len(items))
	for _, doc := range items {
		uploader := "Unknown"
		if doc.UploadedBy != nil {
			uploader = h.getUserName(*doc.UploadedBy)
		}
		out = append(out, IssueAttachmentDTO{
			ID:               doc.ID,
			DocumentType:     string(doc.DocumentType),
			OriginalFileName: doc.OriginalFileName,
			MimeType:         doc.MimeType,
			CVStatus:         string(doc.CVStatus),
			CVConfidence:     doc.CVConfidence,
			CreatedAt:        doc.CreatedAt,
			UploadedBy:       uploader,
			DownloadURL:      fmt.Sprintf("/api/storage/download/%d", doc.ID),
		})
	}
	return out
}

func (h *Handler) loadCommentUserNames(items []models.IssueComment) map[uint]string {
	ids := make([]uint, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.AuthorID)
	}
	return h.loadUserNames(ids)
}

func (h *Handler) loadUserNamesFromHistory(items []models.IssueStatusHistory) map[uint]string {
	ids := make([]uint, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ChangedByID)
	}
	return h.loadUserNames(ids)
}

func (h *Handler) loadUserNames(ids []uint) map[uint]string {
	out := make(map[uint]string)
	if len(ids) == 0 {
		return out
	}
	var users []models.User
	if err := h.db.Where("id IN ?", ids).Find(&users).Error; err == nil {
		for _, u := range users {
			out[u.ID] = u.FullName
		}
	}
	return out
}

func (h *Handler) getUserName(userID uint) string {
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		return "Unknown"
	}
	return user.FullName
}

func (h *Handler) isOverdue(issue models.Issue) bool {
	if issue.DueDate == nil {
		return false
	}
	if issue.Status == models.IssueStatusAccepted {
		return false
	}
	return issue.DueDate.Before(time.Now())
}

func (h *Handler) displayStatus(issue models.Issue) string {
	if h.isOverdue(issue) {
		return "overdue"
	}
	return string(issue.Status)
}
