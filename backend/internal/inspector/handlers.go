package inspector

import (
	//"errors"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/backend/internal/auth"
	"github.com/Loc-Leonard/Diploma/backend/internal/cv"
	"github.com/Loc-Leonard/Diploma/backend/internal/models"
	"github.com/Loc-Leonard/Diploma/backend/internal/objectcore"
)

type Handler struct {
	db          *gorm.DB
	cvProcessor cv.HTTPProcessor
	storageRoot string
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB, cvProcessor cv.HTTPProcessor, storageRoot string) {
	h := &Handler{
		db:          db,
		cvProcessor: cvProcessor,
		storageRoot: storageRoot,
	}

	gr := r.Group("/inspector")
	gr.Use(
		auth.AuthMiddleware(),
		auth.MustChangePasswordMiddleware(db),
		auth.InspectorOnly(),
	)
	{
		gr.GET("/dashboard/checks", h.DashboardChecks)
		gr.GET("/dashboard/objects", h.DashboardObjects)

		gr.GET("/objects", h.ObjectsList)
		gr.GET("/objects/:id", h.ObjectDetails)
		gr.POST("/objects/:id/activation-decision", h.ActivationDecision)

		// Document management endpoints
		gr.GET("/objects/:id/documents", h.ListDocuments)
		gr.POST("/objects/:id/documents/upload", h.UploadDocument)
		gr.DELETE("/objects/:id/documents/:docId", h.DeleteDocument)
	}
}

// GET /inspector/dashboard/checks
func (h *Handler) DashboardChecks(c *gin.Context) {
	inspectorID := auth.UserIDFromContext(c)
	if inspectorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var inspections []models.Inspection

	tx := h.db.
		Preload("Object").
		Where("inspector_id = ?", inspectorID)

	if status := c.Query("status"); status != "" {
		tx = tx.Where("inspections.status = ?", status)
	}
	if city := c.Query("city"); city != "" {
		tx = tx.Joins("JOIN objects ON objects.id = inspections.object_id").
			Where("objects.city = ?", city)
	}

	if err := tx.Order("planned_at ASC").Find(&inspections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load checks"})
		return
	}

	out := make([]InspectorDashboardCheck, 0, len(inspections))
	for _, ins := range inspections {
		out = append(out, InspectorDashboardCheck{
			ID:         ins.ID,
			ObjectID:   ins.ObjectID,
			ObjectName: ins.Object.Name,
			City:       ins.Object.City,
			Address:    ins.Object.Address,
			Status:     string(ins.Status),
			PlannedAt:  ins.PlannedAt,
			IssuesOpen: ins.IssuesOpen,
		})
	}

	c.JSON(http.StatusOK, out)
}

// GET /inspector/dashboard/objects
func (h *Handler) DashboardObjects(c *gin.Context) {
	inspectorID := auth.UserIDFromContext(c)
	if inspectorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	type row struct {
		ID            uint
		Name          string
		City          string
		Address       string
		ForemanName   string
		ActiveChecks  int
		OverdueChecks int
		OpenIssues    int
	}

	var rows []row
	if err := h.db.Raw(`
		SELECT
			o.id,
			o.name,
			o.city,
			o.address,
			COALESCE(u.full_name, '') AS foreman_name,
			SUM(CASE WHEN i.status IN ('PLANNED','IN_PROGRESS') THEN 1 ELSE 0 END) AS active_checks,
			SUM(CASE WHEN i.status = 'OVERDUE' THEN 1 ELSE 0 END) AS overdue_checks,
			SUM(i.issues_open) AS open_issues
		FROM inspections i
		JOIN objects o ON o.id = i.object_id
		LEFT JOIN users u ON u.id = o.foreman_user_id
		WHERE i.inspector_id = ?
		GROUP BY o.id, o.name, o.city, o.address, u.full_name
		ORDER BY o.name
	`, inspectorID).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load objects"})
		return
	}

	out := make([]InspectorDashboardObject, 0, len(rows))
	for _, r := range rows {
		out = append(out, InspectorDashboardObject{
			ID:            r.ID,
			Name:          r.Name,
			City:          r.City,
			Address:       r.Address,
			ForemanName:   r.ForemanName,
			ActiveChecks:  r.ActiveChecks,
			OverdueChecks: r.OverdueChecks,
			OpenIssues:    r.OpenIssues,
		})
	}

	c.JSON(http.StatusOK, out)
}

// GET /inspector/objects
func (h *Handler) ObjectsList(c *gin.Context) {
	inspectorID := auth.UserIDFromContext(c)
	if inspectorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	type row struct {
		ID               uint
		Name             string
		City             string
		Address          string
		Status           models.ObjectStatus
		ForemanName      string
		PlannedStartDate *time.Time
		Lat              float64
		Lng              float64
	}

	var rows []row

	tx := h.db.Raw(`
		SELECT
			o.id,
			o.name,
			o.city,
			o.address,
			o.status,
			COALESCE(u.full_name, '') AS foreman_name,
			o.planned_start_date,
			o.lat,
			o.lng
		FROM objects o
		LEFT JOIN users u ON u.id = o.foreman_user_id
		WHERE o.inspector_user_id = ?
		ORDER BY
			CASE
				WHEN o.status = 'WAITING_INSPECTOR_CONFIRMATION' THEN 0
				WHEN o.status = 'ACTIVE' THEN 1
				ELSE 2
			END,
			o.id DESC
	`, inspectorID)

	if status := c.Query("status"); status != "" {
		tx = h.db.Raw(`
			SELECT
				o.id,
				o.name,
				o.city,
				o.address,
				o.status,
				COALESCE(u.full_name, '') AS foreman_name,
				o.planned_start_date,
				o.lat,
				o.lng
			FROM objects o
			LEFT JOIN users u ON u.id = o.foreman_user_id
			WHERE o.inspector_user_id = ? AND o.status = ?
			ORDER BY o.id DESC
		`, inspectorID, status)
	}

	if err := tx.Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load objects"})
		return
	}

	out := make([]InspectorObjectListItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, InspectorObjectListItem{
			ID:               r.ID,
			Name:             r.Name,
			City:             r.City,
			Address:          r.Address,
			Status:           r.Status,
			ForemanName:      r.ForemanName,
			PlannedStartDate: r.PlannedStartDate,
			Lat:              r.Lat,
			Lng:              r.Lng,
			Progress:         objectcore.CalcProgress(h.db, r.ID),
			HasPendingAction: r.Status == models.ObjectStatusWaitingInspectorConfirmation,
		})
	}

	c.JSON(http.StatusOK, out)
}

// GET /inspector/objects/:id
func (h *Handler) ObjectDetails(c *gin.Context) {
	inspectorID := auth.UserIDFromContext(c)
	if inspectorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")

	var obj models.Object
	if err := h.db.
		Where("id = ? AND inspector_user_id = ?", id, inspectorID).
		First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}

	// Загружаем связанных пользователей
	var customer, foreman, inspector models.User

	customerName := ""
	if obj.CustomerControlUserID != 0 {
		if err := h.db.First(&customer, obj.CustomerControlUserID).Error; err == nil {
			customerName = customer.FullName
		}
	}

	foremanName := ""
	if obj.ForemanUserID != 0 {
		if err := h.db.First(&foreman, obj.ForemanUserID).Error; err == nil {
			foremanName = foreman.FullName
		}
	}

	inspectorName := ""
	if obj.InspectorUserID != 0 {
		if err := h.db.First(&inspector, obj.InspectorUserID).Error; err == nil {
			inspectorName = inspector.FullName
		}
	}

	// Загружаем work_items и deliveries
	var workItems []models.WorkItem
	h.db.Where("object_id = ?", obj.ID).Order("id ASC").Find(&workItems)

	var deliveries []models.MaterialDelivery
	h.db.Where("object_id = ?", obj.ID).Order("id ASC").Find(&deliveries)

	objDTO := ObjectDTO{
		ID:                     obj.ID,
		Name:                   obj.Name,
		City:                   obj.City,
		Address:                obj.Address,
		Description:            obj.Description,
		Status:                 obj.Status,
		Lat:                    obj.Lat,
		Lng:                    obj.Lng,
		PlannedStartDate:       obj.PlannedStartDate,
		PlannedEndDate:         obj.PlannedEndDate,
		ActualStartDate:        obj.ActualStartDate,
		InitChecklistJSON:      obj.InitChecklistJSON,
		InitActFilePath:        obj.InitActFilePath,
		ActivationRejectReason: obj.ActivationRejectReason,
	}

	if customerName != "" {
		objDTO.Customer = &PersonDTO{ID: obj.CustomerControlUserID, FullName: customerName}
	}
	if foremanName != "" {
		objDTO.Foreman = &PersonDTO{ID: obj.ForemanUserID, FullName: foremanName}
	}
	if inspectorName != "" {
		objDTO.Inspector = &PersonDTO{ID: obj.InspectorUserID, FullName: inspectorName}
	}

	c.JSON(http.StatusOK, gin.H{
		"object":     objDTO,
		"work_items": workItems,
		"deliveries": deliveries,
	})
}

// POST /inspector/objects/:id/activation-decision
func (h *Handler) ActivationDecision(c *gin.Context) {
	inspectorID := auth.UserIDFromContext(c)
	if inspectorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")

	var obj models.Object
	if err := h.db.
		Where("id = ? AND inspector_user_id = ?", id, inspectorID).
		First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}

	if obj.Status != models.ObjectStatusWaitingInspectorConfirmation {
		c.JSON(http.StatusBadRequest, gin.H{"error": "object is not waiting for confirmation"})
		return
	}

	var req ActivationDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	decision := strings.ToUpper(strings.TrimSpace(req.Decision))

	switch decision {
	case "APPROVE":
		obj.Status = models.ObjectStatusActive
		obj.ActualStartDate = &now
		obj.ActivationRejectReason = ""
		obj.ActivationReviewedAt = &now

	case "REJECT":
		reason := strings.TrimSpace(req.RejectionReason)
		if reason == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rejection_reason is required"})
			return
		}
		obj.Status = models.ObjectStatusPlanned
		obj.ActualStartDate = nil
		obj.ActivationRejectReason = reason
		obj.ActivationReviewedAt = &now

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid decision"})
		return
	}

	if err := h.db.Save(&obj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"decision": decision,
	})
}

// GET /inspector/objects/:id/documents
func (h *Handler) ListDocuments(c *gin.Context) {
	inspectorID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	// Проверяем доступ к объекту
	var obj models.Object
	if err := h.db.Where("id = ? AND inspector_user_id = ?", objectID, inspectorID).
		First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	// Получаем документы, связанные с объектом
	var documents []models.MaterialDocument
	if err := h.db.
		Joins("LEFT JOIN material_deliveries ON material_deliveries.id = material_documents.delivery_id").
		Where("material_deliveries.object_id = ? OR material_documents.storage_path LIKE ?", obj.ID, "%/"+objectID+"/%").
		Order("material_documents.created_at DESC").
		Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	// Загружаем информацию о пользователях
	userIDs := make([]uint, 0)
	for _, doc := range documents {
		if doc.UploadedBy != nil {
			userIDs = append(userIDs, *doc.UploadedBy)
		}
		if doc.DeliveryID != nil {
			var delivery models.MaterialDelivery
			if err := h.db.First(&delivery, *doc.DeliveryID).Error; err == nil {
				if delivery.ForemanID != 0 {
					userIDs = append(userIDs, delivery.ForemanID)
				}
			}
		}
	}

	usersMap := make(map[uint]string)
	if len(userIDs) > 0 {
		var users []models.User
		if err := h.db.Where("id IN ?", userIDs).Find(&users).Error; err == nil {
			for _, u := range users {
				usersMap[u.ID] = u.FullName
			}
		}
	}

	resp := make([]DocumentDTO, 0, len(documents))
	for _, doc := range documents {
		uploadedBy := "Unknown"

		// Сначала пробуем получить из UploadedBy
		if doc.UploadedBy != nil {
			if name, ok := usersMap[*doc.UploadedBy]; ok {
				uploadedBy = name
			}
		}

		// Если нет, пробуем из delivery
		if uploadedBy == "Unknown" && doc.DeliveryID != nil {
			var delivery models.MaterialDelivery
			if err := h.db.First(&delivery, *doc.DeliveryID).Error; err == nil {
				if delivery.ForemanID != 0 {
					if name, ok := usersMap[delivery.ForemanID]; ok {
						uploadedBy = name
					}
				}
			}
		}

		// Извлекаем CV confidence из payload
		cvConfidence := doc.CVConfidence
		if cvConfidence == 0 && doc.CVPayloadJSON != "" {
			var cvData map[string]interface{}
			if err := json.Unmarshal([]byte(doc.CVPayloadJSON), &cvData); err == nil {
				if extraction, ok := cvData["extraction"].(map[string]interface{}); ok {
					if conf, ok := extraction["confidence"].(float64); ok {
						cvConfidence = conf
					}
				}
			}
		}

		resp = append(resp, DocumentDTO{
			ID:               doc.ID,
			DocumentType:     string(doc.DocumentType),
			OriginalFileName: doc.OriginalFileName,
			MimeType:         doc.MimeType,
			CVStatus:         string(doc.CVStatus),
			CVConfidence:     cvConfidence,
			CreatedAt:        doc.CreatedAt,
			UploadedBy:       uploadedBy,
			DownloadURL:      fmt.Sprintf("/api/storage/download/%d", doc.ID),
		})
	}

	c.JSON(http.StatusOK, resp)
}

// POST /inspector/objects/:id/documents/upload
func (h *Handler) UploadDocument(c *gin.Context) {
	inspectorID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	// Проверяем доступ к объекту
	var obj models.Object
	if err := h.db.Where("id = ? AND inspector_user_id = ?", objectID, inspectorID).
		First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "no file provided"})
		return
	}

	// Проверка размера файла (макс 10MB)
	const maxFileSize = 10 * 1024 * 1024
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "file too large (max 10MB)"})
		return
	}

	// Проверка типа файла
	allowedMimeTypes := []string{
		"image/jpeg", "image/png", "image/jpg", "image/gif",
		"application/pdf",
		"application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}

	fileHeader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "cannot read file"})
		return
	}
	defer fileHeader.Close()

	buffer := make([]byte, 512)
	_, err = fileHeader.Read(buffer)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "cannot read file"})
		return
	}
	fileType := http.DetectContentType(buffer)

	if !slices.Contains(allowedMimeTypes, fileType) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "file type not allowed"})
		return
	}

	// Определяем тип документа из формы или используем дефолт
	docTypeStr := c.PostForm("document_type")
	if docTypeStr == "" {
		docTypeStr = "OTHER"
	}

	// Создаем директорию для объекта
	objectDir := filepath.Join(h.storageRoot, "inspector", objectID, "documents")
	if err := os.MkdirAll(objectDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "cannot create directory"})
		return
	}

	// Генерируем уникальное имя файла
	timestamp := time.Now().Format("20060102_150405")
	originalFilename := filepath.Base(file.Filename)
	// Очищаем имя файла от специальных символов
	cleanFilename := strings.ReplaceAll(originalFilename, " ", "_")
	filename := timestamp + "_" + cleanFilename
	filePath := filepath.Join(objectDir, filename)

	// Сохраняем файл
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "cannot save file"})
		return
	}

	// Создаем запись о документе СНАЧАЛА (чтобы получить ID для горотины)
	doc := models.MaterialDocument{
		DeliveryID:       nil,
		UploadedBy:       &inspectorID,
		DocumentType:     models.MaterialDocumentType(docTypeStr),
		StoragePath:      filePath,
		OriginalFileName: originalFilename,
		MimeType:         fileType,
		CVStatus:         models.CVProcessingStatusPending,
		CVPayloadJSON:    "",
		CVConfidence:     0.0,
	}

	if err := h.db.Create(&doc).Error; err != nil {
		os.Remove(filePath) // Откат - удаляем файл при ошибке БД
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error: document"})
		return
	}

	// Сохраняем ID ДО запуска горотины
	documentID := doc.ID

	// Запускаем CV обработку в горутине (асинхронно)
	go func(docID uint, fPath string, fName string) {
		ctx := context.Background()
		cvResult, cvErr := h.cvProcessor.ProcessFile(ctx, fPath, fName)

		updateStatus := models.CVProcessingStatusDone
		updatePayload := ""
		updateConfidence := 0.0
		updateDocType := ""

		if cvErr != nil {
			log.Printf("CV processing failed for file %s: %v", fName, cvErr)
			updateStatus = models.CVProcessingStatusFailed
		} else {
			updatePayload = string(cvResult.RawJSON)
			updateConfidence = cvResult.Extraction.Confidence

			// Обновляем тип документа на основе CV
			if cvResult.Extraction.DocumentType != "" {
				switch cvResult.Extraction.DocumentType {
				case "TTN":
					updateDocType = "TTN"
				case "QUALITY_PASSPORT":
					updateDocType = "QUALITY_PASSPORT"
				case "PHOTO":
					updateDocType = "PHOTO"
				}
			}
		}

		// Обновляем запись в БД
		updates := map[string]interface{}{
			"cv_status":       updateStatus,
			"cv_payload_json": updatePayload,
			"cv_confidence":   updateConfidence,
		}
		if updateDocType != "" {
			updates["document_type"] = updateDocType
		}

		if err := h.db.Model(&models.MaterialDocument{}).Where("id = ?", docID).Updates(updates).Error; err != nil {
			log.Printf("Failed to update CV status for document %d: %v", docID, err)
		}
	}(documentID, filePath, originalFilename)

	// Отвечаем клиенту сразу
	c.JSON(http.StatusCreated, DocumentUploadResponse{
		Status:     "uploaded",
		DocumentID: doc.ID,
		FileName:   originalFilename,
		FilePath:   filePath,
		CVStatus:   string(models.CVProcessingStatusPending),
	})
}

// DELETE /inspector/objects/:id/documents/:docId
func (h *Handler) DeleteDocument(c *gin.Context) {
	inspectorID := auth.UserIDFromContext(c)
	objectID := c.Param("id")
	docID := c.Param("docId")

	// Проверяем доступ к объекту
	var obj models.Object
	if err := h.db.Where("id = ? AND inspector_user_id = ?", objectID, inspectorID).
		First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	var doc models.MaterialDocument
	if err := h.db.First(&doc, docID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "document not found"})
		return
	}

	// Проверяем, что документ принадлежит этому объекту
	expectedPath := filepath.Join(h.storageRoot, "inspector", objectID, "documents")
	if !strings.HasPrefix(doc.StoragePath, expectedPath) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "document does not belong to this object"})
		return
	}

	// Проверяем, не связан ли документ с delivery
	if doc.DeliveryID != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "cannot delete document linked to delivery"})
		return
	}

	// Удаляем файл с диска
	if err := os.Remove(doc.StoragePath); err != nil {
		log.Printf("Failed to delete file %s: %v", doc.StoragePath, err)
		// Продолжаем удаление записи из БД
	}

	// Удаляем запись из БД
	if err := h.db.Delete(&doc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// for tests
func HandlerForTest(db *gorm.DB) *Handler {
	return &Handler{db: db}
}
