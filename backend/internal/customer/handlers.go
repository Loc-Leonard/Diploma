package customer

import (
	"context"
	"encoding/json"
	"errors"
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

// Route paths
const (
	RouteCustomerDashboardObjects = "/customer/dashboard/objects"
	RouteCustomerDashboardForemen = "/customer/dashboard/foremen"
	RouteCustomerForemenList      = "/customer/foremen-list"
	RouteCustomerInspectorsList   = "/customer/inspectors-list"
	RouteCustomerObjectsCreate    = "/customer/objects"
	RouteCustomerObjectsActivate  = "/customer/objects/:id/activate"
	RouteCustomerObjectsGet       = "/customer/objects/:id"
	RouteCustomerWorkItemsList    = "/customer/objects/:id/work-items"
	RouteCustomerWorkItemsCreate  = "/customer/objects/:id/work-items"
	RouteCustomerWorkItemsUpdate  = "/customer/objects/:id/work-items/:wid"
	RouteCustomerWorkItemsDelete  = "/customer/objects/:id/work-items/:wid"
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

	gr := r.Group("/customer")
	gr.Use(auth.AuthMiddleware(), auth.MustChangePasswordMiddleware(db), auth.CustomerOnly())
	{
		gr.GET("/dashboard/objects", h.DashboardObjects)
		gr.GET("/dashboard/foremen", h.DashboardForemen)

		gr.GET("/foremen-list", h.ForemenList)
		gr.GET("/inspectors-list", h.InspectorsList)

		gr.POST("/objects", h.CreateObject)
		gr.POST("/objects/:id/activate", h.ActivateObject)
		gr.GET("/objects/:id", h.GetObject)

		// Этдпоинты для работы с просмотром видов работ
		gr.GET("/objects/:id/work-items", h.ListWorkItems)
		gr.POST("/objects/:id/work-items", h.CreateWorkItem)
		gr.PUT("/objects/:id/work-items/:wid", h.UpdateWorkItem)
		gr.DELETE("/objects/:id/work-items/:wid", h.DeleteWorkItem)

		// Эндпоинты для работы с документами (CV)
		gr.GET("/objects/:id/documents", h.ListDocuments)
		gr.POST("/objects/:id/documents/upload", h.UploadDocument)
		gr.DELETE("/objects/:id/documents/:docId", h.DeleteDocument)
	}
}

func (h *Handler) ForemenList(c *gin.Context) {
	var users []models.User
	if err := h.db.
		Where("role = ?", models.RoleForeman).
		Order("full_name ASC").
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	resp := make([]models.SimpleUserDTO, 0, len(users))
	for _, u := range users {
		resp = append(resp, models.SimpleUserDTO{
			ID:       u.ID,
			FullName: u.FullName,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) InspectorsList(c *gin.Context) {
	var users []models.User
	if err := h.db.
		Where("role = ?", models.RoleInspector).
		Order("full_name ASC").
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	resp := make([]models.SimpleUserDTO, 0, len(users))
	for _, u := range users {
		resp = append(resp, models.SimpleUserDTO{
			ID:       u.ID,
			FullName: u.FullName,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateObject(c *gin.Context) {
	customerID := auth.UserIDFromContext(c)

	var req CreateObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	var inspector models.User
	if err := h.db.
		Where("id = ? AND role = ?", req.InspectorUserID, models.RoleInspector).
		First(&inspector).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid inspector"})
		return
	}

	obj := models.Object{
		Name:                  req.Name,
		Address:               req.Address,
		City:                  req.City,
		Description:           req.Description,
		Status:                models.ObjectStatusPlanned,
		Lat:                   req.Lat,
		Lng:                   req.Lng,
		CustomerControlUserID: customerID,
		ForemanUserID:         req.ForemanUserID,
		InspectorUserID:       req.InspectorUserID,
		PlannedStartDate:      req.PlannedStartDate,
		PlannedEndDate:        req.PlannedEndDate,
	}

	if err := h.db.Create(&obj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}
	c.JSON(http.StatusCreated, models.IDResponse{ID: obj.ID})
}

func (h *Handler) DashboardObjects(c *gin.Context) {
	userID := auth.UserIDFromContext(c)

	var objects []models.Object
	q := h.db.Where("customer_control_user_id = ?", userID)

	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	if city := c.Query("city"); city != "" {
		q = q.Where("city = ?", city)
	}
	if err := q.Order("id ASC").Find(&objects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	foremanIDs := make([]uint, 0)
	for _, o := range objects {
		if o.ForemanUserID != 0 {
			foremanIDs = append(foremanIDs, o.ForemanUserID)
		}
	}

	foremenMap := map[uint]models.User{}
	if len(foremanIDs) > 0 {
		var foremen []models.User
		if err := h.db.
			Where("id IN ?", foremanIDs).
			Where("role = ?", models.RoleForeman).
			Find(&foremen).Error; err == nil {
			for _, f := range foremen {
				foremenMap[f.ID] = f
			}
		}
	}

	resp := make([]DashboardObjectDTO, 0, len(objects))
	for _, o := range objects {
		dto := DashboardObjectDTO{
			ID:                     o.ID,
			Name:                   o.Name,
			City:                   o.City,
			Address:                o.Address,
			Status:                 o.Status,
			Progress:               objectcore.CalcProgress(h.db, o.ID),
			PlannedStartDate:       o.PlannedStartDate,
			PlannedEndDate:         o.PlannedEndDate,
			Lat:                    o.Lat,
			Lng:                    o.Lng,
			ActivationRejectReason: o.ActivationRejectReason,
		}

		if f, ok := foremenMap[o.ForemanUserID]; ok {
			dto.Foreman = &struct {
				ID       uint   `json:"id"`
				FullName string `json:"full_name"`
			}{
				ID:       f.ID,
				FullName: f.FullName,
			}
		}

		resp = append(resp, dto)
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) DashboardForemen(c *gin.Context) {
	userID := auth.UserIDFromContext(c)

	var objects []models.Object
	if err := h.db.
		Where("customer_control_user_id = ?", userID).
		Where("foreman_user_id IS NOT NULL").
		Find(&objects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	foremanObject := make(map[uint]models.Object)
	foremanIDs := make([]uint, 0)
	for _, o := range objects {
		if o.ForemanUserID == 0 {
			continue
		}
		if _, ok := foremanObject[o.ForemanUserID]; !ok {
			foremanObject[o.ForemanUserID] = o
			foremanIDs = append(foremanIDs, o.ForemanUserID)
		}
	}

	if len(foremanIDs) == 0 {
		c.JSON(http.StatusOK, []DashboardForemanDTO{})
		return
	}

	var foremen []models.User
	if err := h.db.
		Where("id IN ?", foremanIDs).
		Where("role = ?", models.RoleForeman).
		Find(&foremen).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	resp := make([]DashboardForemanDTO, 0, len(foremen))
	for _, f := range foremen {
		obj := foremanObject[f.ID]
		dto := DashboardForemanDTO{
			ID:       f.ID,
			FullName: f.FullName,
			City:     obj.City,
			CurrentObject: &struct {
				ID   uint   `json:"id"`
				Name string `json:"name"`
			}{
				ID:   obj.ID,
				Name: obj.Name,
			},
		}
		resp = append(resp, dto)
	}

	c.JSON(http.StatusOK, resp)
}

// POST /customer/objects/:id/activate
func (h *Handler) ActivateObject(c *gin.Context) {
	customerID := auth.UserIDFromContext(c)
	if customerID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized"})
		return
	}

	id := c.Param("id")

	var obj models.Object
	if err := h.db.
		Where("id = ? AND customer_control_user_id = ?", id, customerID).
		First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	if obj.Status != models.ObjectStatusPlanned {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "object is not in PLANNED status"})
		return
	}

	var req ActivateObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	obj.InitChecklistJSON = req.ChecklistJSON

	if req.ActFilePath != nil {
		obj.InitActFilePath = *req.ActFilePath
	}

	obj.ActivationRejectReason = ""
	obj.ActivationReviewedAt = nil
	obj.Status = models.ObjectStatusWaitingInspectorConfirmation

	if err := h.db.Save(&obj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Status: "sent_for_inspector_confirmation"})
}

// GET /customer/objects/:id
func (h *Handler) GetObject(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	id := c.Param("id")

	obj, err := objectcore.LoadObjectForUser(h.db, id, userID, string(models.RoleCustomer))
	if errors.Is(err, objectcore.ErrNotFound) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}
	if errors.Is(err, objectcore.ErrForbidden) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "access denied"})
		return
	}
	detail := objectcore.BuildObjectDetailDTO(h.db, obj)
	c.JSON(http.StatusOK, detail)
}

func (h *Handler) ListWorkItems(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	var obj models.Object
	if err := h.db.Where("id = ? AND customer_control_user_id = ?", objectID, userID).
		First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	var items []models.WorkItem
	h.db.Where("object_id = ?", obj.ID).Order("sort_order ASC, id ASC").Find(&items)
	c.JSON(http.StatusOK, items)
}

func (h *Handler) CreateWorkItem(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	var obj models.Object
	if err := h.db.Where("id = ? AND customer_control_user_id = ?", objectID, userID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	var in WorkItemInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	item := models.WorkItem{
		ObjectID:         obj.ID,
		Name:             in.Name,
		Description:      in.Description,
		Unit:             in.Unit,
		PlanQty:          in.PlanQty,
		PlannedStartDate: in.PlannedStartDate,
		PlannedEndDate:   in.PlannedEndDate,
		SortOrder:        in.SortOrder,
		Status:           models.WorkItemStatusPlanned,
		DependsOnID:      in.DependsOnID,
	}

	if err := h.db.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) UpdateWorkItem(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")
	wid := c.Param("wid")

	var obj models.Object
	if err := h.db.Where("id = ? AND customer_control_user_id = ?", objectID, userID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	var item models.WorkItem
	if err := h.db.Where("id = ? AND object_id = ?", wid, obj.ID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "work item not found"})
		return
	}

	var in WorkItemInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	item.Name = in.Name
	item.Description = in.Description
	item.Unit = in.Unit
	item.PlanQty = in.PlanQty
	item.PlannedStartDate = in.PlannedStartDate
	item.PlannedEndDate = in.PlannedEndDate
	item.SortOrder = in.SortOrder
	item.DependsOnID = in.DependsOnID

	if err := h.db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteWorkItem(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")
	wid := c.Param("wid")

	var obj models.Object
	if err := h.db.Where("id = ? AND customer_control_user_id = ?", objectID, userID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "object not found"})
		return
	}

	var item models.WorkItem
	if err := h.db.Where("id = ? AND object_id = ?", wid, obj.ID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "work item not found"})
		return
	}

	if err := h.db.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "work item deleted"})
}

// GET /customer/objects/:id/documents
func (h *Handler) ListDocuments(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	var obj models.Object
	if err := h.db.Where("id = ? AND customer_control_user_id = ?", objectID, userID).
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

// POST /customer/objects/:id/documents/upload
func (h *Handler) UploadDocument(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	var obj models.Object
	if err := h.db.Where("id = ? AND customer_control_user_id = ?", objectID, userID).
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
	objectDir := filepath.Join(h.storageRoot, "customer", objectID, "documents")
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
		UploadedBy:       &userID, // 👈 Сохраняем кто загрузил
		DocumentType:     models.MaterialDocumentType(docTypeStr),
		StoragePath:      filePath,
		OriginalFileName: originalFilename,
		MimeType:         fileType,
		CVStatus:         models.CVProcessingStatusPending,
		CVPayloadJSON:    "",
	}

	if err := h.db.Create(&doc).Error; err != nil {
		os.Remove(filePath) // Откат - удаляем файл при ошибке БД
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "db error: document"})
		return
	}

	// 👈 Теперь сохраняем ID ДО запуска горотины
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

// DELETE /customer/objects/:id/documents/:docId
func (h *Handler) DeleteDocument(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")
	docID := c.Param("docId")

	var obj models.Object
	if err := h.db.Where("id = ? AND customer_control_user_id = ?", objectID, userID).
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
	expectedPath := filepath.Join(h.storageRoot, "customer", objectID, "documents")
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
