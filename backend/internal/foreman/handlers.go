package foreman

import (
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

	gr := r.Group("/foreman")
	gr.Use(
		auth.AuthMiddleware(),
		auth.MustChangePasswordMiddleware(db),
		auth.ForemanOnly(),
	)
	{
		gr.GET("/objects", h.ObjectsList)
		gr.GET("/objects/:id", h.ObjectDetails)

		// Work reports
		gr.POST("/objects/:id/work-reports", h.SubmitWorkReports)

		// Deliveries
		gr.GET("/objects/:id/deliveries", h.ListDeliveries)
		gr.POST("/objects/:id/deliveries", h.CreateDelivery)

		// Documents (CV)
		gr.GET("/objects/:id/documents", h.ListDocuments)
		gr.POST("/objects/:id/documents/upload", h.UploadDocument)
		gr.DELETE("/objects/:id/documents/:docId", h.DeleteDocument)
	}
}

// GET /foreman/objects
func (h *Handler) ObjectsList(c *gin.Context) {
	foremanID := auth.UserIDFromContext(c)
	if foremanID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized"})
		return
	}

	var objects []models.Object
	if err := h.db.Where("foreman_user_id = ?", foremanID).Order("id DESC").Find(&objects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	resp := make([]ForemanObjectDTO, 0, len(objects))
	for _, o := range objects {
		resp = append(resp, ForemanObjectDTO{
			ID:               o.ID,
			Name:             o.Name,
			City:             o.City,
			Address:          o.Address,
			Status:           o.Status,
			PlannedStartDate: o.PlannedStartDate,
			PlannedEndDate:   o.PlannedEndDate,
		})
	}

	c.JSON(http.StatusOK, resp)
}

// GET /foreman/objects/:id
func (h *Handler) ObjectDetails(c *gin.Context) {
	foremanID := auth.UserIDFromContext(c)
	if foremanID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized"})
		return
	}

	id := c.Param("id")

	var obj models.Object
	if err := h.db.Where("id = ? AND foreman_user_id = ?", id, foremanID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	var workItems []models.WorkItem
	h.db.Where("object_id = ?", obj.ID).Order("id ASC").Find(&workItems)

	var deliveries []models.MaterialDelivery
	h.db.Where("object_id = ?", obj.ID).Order("id ASC").Find(&deliveries)

	// Преобразуем WorkItem в WorkItemDTO
	workItemDTOs := make([]WorkItemDTO, 0, len(workItems))
	for _, wi := range workItems {
		workItemDTOs = append(workItemDTOs, WorkItemDTO{
			ID:               wi.ID,
			ObjectID:         wi.ObjectID,
			Name:             wi.Name,
			Description:      wi.Description,
			Unit:             wi.Unit,
			PlanQty:          wi.PlanQty,
			PlannedStartDate: wi.PlannedStartDate,
			PlannedEndDate:   wi.PlannedEndDate,
			SortOrder:        wi.SortOrder,
			Status:           string(wi.Status),
		})
	}

	// Преобразуем MaterialDelivery в DeliveryDTO
	deliveryDTOs := make([]DeliveryDTO, 0, len(deliveries))
	for _, d := range deliveries {
		deliveryDTOs = append(deliveryDTOs, DeliveryDTO{
			ID:             d.ID,
			ObjectID:       d.ObjectID,
			ForemanID:      d.ForemanID,
			Date:           d.Date,
			Material:       d.Material,
			Qty:            d.Qty,
			Unit:           d.Unit,
			DocumentNumber: d.DocumentNumber,
			Source:         d.Source,
			CVConfidence:   d.CVConfidence,
		})
	}

	c.JSON(http.StatusOK, ObjectDetailResponse{
		Object:     buildObjectDTO(obj),
		WorkItems:  workItemDTOs,
		Deliveries: deliveryDTOs,
	})
}

func buildObjectDTO(obj models.Object) ObjectDTO {
	return ObjectDTO{
		ID:               obj.ID,
		Name:             obj.Name,
		City:             obj.City,
		Address:          obj.Address,
		Description:      obj.Description,
		Status:           obj.Status,
		Lat:              obj.Lat,
		Lng:              obj.Lng,
		PlannedStartDate: obj.PlannedStartDate,
		PlannedEndDate:   obj.PlannedEndDate,
		ActualStartDate:  obj.ActualStartDate,
	}
}

// POST /foreman/objects/:id/work-reports
func (h *Handler) SubmitWorkReports(c *gin.Context) {
	foremanID := auth.UserIDFromContext(c)
	if foremanID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized"})
		return
	}

	id := c.Param("id")

	var obj models.Object
	if err := h.db.Where("id = ? AND foreman_user_id = ?", id, foremanID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	var req struct {
		Reports []struct {
			WorkItemID uint   `json:"work_item_id"`
			Qty        int    `json:"qty"`
			Date       string `json:"date"`
		} `json:"reports"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if len(req.Reports) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "no reports provided"})
		return
	}

	for _, r := range req.Reports {
		// Парсим дату из строки
		date, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			// Пробуем другой формат с временем
			date, err = time.Parse(time.RFC3339, r.Date)
			if err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid date format"})
				return
			}
		}

		report := models.WorkReport{
			WorkItemID: r.WorkItemID,
			ObjectID:   obj.ID,
			ForemanID:  foremanID,
			Qty:        float64(r.Qty),
			Date:       date,
			Status:     models.WorkReportStatusSubmitted,
		}
		if err := h.db.Create(&report).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
			return
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Status: "reports submitted"})
}

// GET /foreman/objects/:id/deliveries
func (h *Handler) ListDeliveries(c *gin.Context) {
	foremanID := auth.UserIDFromContext(c)
	if foremanID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized"})
		return
	}

	id := c.Param("id")

	var obj models.Object
	if err := h.db.Where("id = ? AND foreman_user_id = ?", id, foremanID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	var deliveries []models.MaterialDelivery
	if err := h.db.Where("object_id = ? AND foreman_id = ?", obj.ID, foremanID).Order("id DESC").Find(&deliveries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	c.JSON(http.StatusOK, deliveries)
}

// POST /foreman/objects/:id/deliveries
func (h *Handler) CreateDelivery(c *gin.Context) {
	foremanID := auth.UserIDFromContext(c)
	if foremanID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized"})
		return
	}

	id := c.Param("id")

	var obj models.Object
	if err := h.db.Where("id = ? AND foreman_user_id = ?", id, foremanID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	var req struct {
		Material string  `json:"material"`
		Qty      int     `json:"qty"`
		Unit     string  `json:"unit"`
		Date     *string `json:"date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Парсим дату если предоставлена
	var deliveryDate time.Time
	if req.Date != nil {
		var err error
		deliveryDate, err = time.Parse("2006-01-02", *req.Date)
		if err != nil {
			deliveryDate, err = time.Parse(time.RFC3339, *req.Date)
			if err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid date format"})
				return
			}
		}
	} else {
		deliveryDate = time.Now()
	}

	delivery := models.MaterialDelivery{
		ObjectID:     obj.ID,
		ForemanID:    foremanID,
		Material:     req.Material,
		Qty:          float64(req.Qty),
		Unit:         req.Unit,
		Date:         deliveryDate,
		Source:       "MANUAL",
		CVConfidence: 0.0,
	}

	if err := h.db.Create(&delivery).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	c.JSON(http.StatusCreated, delivery)
}

// GET /foreman/objects/:id/documents
func (h *Handler) ListDocuments(c *gin.Context) {
	foremanID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	var obj models.Object
	if err := h.db.Where("id = ? AND foreman_user_id = ?", objectID, foremanID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	var documents []models.MaterialDocument
	if err := h.db.
		Joins("LEFT JOIN material_deliveries ON material_deliveries.id = material_documents.delivery_id").
		Where("material_deliveries.object_id = ? OR material_documents.storage_path LIKE ?", obj.ID, "%/"+objectID+"/%").
		Order("material_documents.created_at DESC").
		Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	userIDs := make([]uint, 0)
	for _, doc := range documents {
		if doc.UploadedBy != nil {
			userIDs = append(userIDs, *doc.UploadedBy)
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
		if doc.UploadedBy != nil {
			if name, ok := usersMap[*doc.UploadedBy]; ok {
				uploadedBy = name
			}
		}

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

// POST /foreman/objects/:id/documents/upload
func (h *Handler) UploadDocument(c *gin.Context) {
	foremanID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	var obj models.Object
	if err := h.db.Where("id = ? AND foreman_user_id = ?", objectID, foremanID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
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
	objectDir := filepath.Join(h.storageRoot, "foreman", objectID, "documents")
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
		UploadedBy:       &foremanID,
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

// DELETE /foreman/objects/:id/documents/:docId
func (h *Handler) DeleteDocument(c *gin.Context) {
	foremanID := auth.UserIDFromContext(c)
	objectID := c.Param("id")
	docID := c.Param("docId")

	// Проверяем доступ к объекту
	var obj models.Object
	if err := h.db.Where("id = ? AND foreman_user_id = ?", objectID, foremanID).
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
	expectedPath := filepath.Join(h.storageRoot, "foreman", objectID, "documents")
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
