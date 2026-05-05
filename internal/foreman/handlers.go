package foreman

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/internal/auth"
	"github.com/Loc-Leonard/Diploma/internal/cv"
	"github.com/Loc-Leonard/Diploma/internal/models"
)

type Handler struct {
	db          *gorm.DB
	cvProcessor cv.FileProcessor
	storageRoot string
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB, cvProcessor cv.FileProcessor, storageRoot string) {
	h := &Handler{
		db:          db,
		cvProcessor: cvProcessor,
		storageRoot: storageRoot,
	}

	gr := r.Group("/foreman")
	gr.Use(
		auth.AuthMiddleware(),
		auth.ForemanOnly(),
		auth.MustChangePasswordMiddleware(db),
	)
	{
		gr.GET("/objects", h.ListObjects)
		gr.GET("/objects/:id", h.ObjectDetail)
		gr.POST("/objects/:id/work-reports", h.CreateWorkReports)
		gr.POST("/objects/:id/deliveries", h.CreateDelivery)
		gr.POST("/objects/:id/deliveries/cv-upload", h.CreateDeliveryFromCV)
	}
}

type ForemanObjectDTO struct {
	ID      uint                `json:"id"`
	Name    string              `json:"name"`
	City    string              `json:"city"`
	Address string              `json:"address"`
	Status  models.ObjectStatus `json:"status"`
}

func (h *Handler) ListObjects(c *gin.Context) {
	foremanID := auth.UserIDFromContext(c)

	var objects []models.Object
	if err := h.db.
		Where("foreman_user_id = ?", foremanID).
		Order("id ASC").
		Find(&objects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	resp := make([]ForemanObjectDTO, 0, len(objects))
	for _, o := range objects {
		resp = append(resp, ForemanObjectDTO{
			ID:      o.ID,
			Name:    o.Name,
			City:    o.City,
			Address: o.Address,
			Status:  o.Status,
		})
	}

	c.JSON(http.StatusOK, resp)
}

type ForemanObjectDetailDTO struct {
	Object     ForemanObjectDTO          `json:"object"`
	WorkItems  []models.WorkItem         `json:"work_items"`
	Deliveries []models.MaterialDelivery `json:"deliveries"`
}

func (h *Handler) ObjectDetail(c *gin.Context) {
	foremanID := auth.UserIDFromContext(c)
	id := c.Param("id")

	var obj models.Object
	if err := h.db.
		Where("id = ? AND foreman_user_id = ?", id, foremanID).
		First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}

	var items []models.WorkItem
	if err := h.db.Where("object_id = ?", obj.ID).Order("id ASC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	var deliveries []models.MaterialDelivery
	if err := h.db.
		Preload("Documents").
		Where("object_id = ?", obj.ID).
		Order("date DESC, id DESC").
		Find(&deliveries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	c.JSON(http.StatusOK, ForemanObjectDetailDTO{
		Object: ForemanObjectDTO{
			ID:      obj.ID,
			Name:    obj.Name,
			City:    obj.City,
			Address: obj.Address,
			Status:  obj.Status,
		},
		WorkItems:  items,
		Deliveries: deliveries,
	})
}

type WorkReportInput struct {
	WorkItemID uint    `json:"work_item_id"`
	Qty        float64 `json:"qty"`
	Date       string  `json:"date"`
}

type WorkReportsRequest struct {
	Reports []WorkReportInput `json:"reports"`
}

func (h *Handler) CreateWorkReports(c *gin.Context) {
	foremanID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	var obj models.Object
	if err := h.db.
		Where("id = ? AND foreman_user_id = ?", objectID, foremanID).
		First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}

	var req WorkReportsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	reports := make([]models.WorkReport, 0, len(req.Reports))
	for _, r := range req.Reports {
		if r.Qty <= 0 {
			continue
		}
		d, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
			return
		}
		reports = append(reports, models.WorkReport{
			ObjectID:   obj.ID,
			WorkItemID: r.WorkItemID,
			ForemanID:  foremanID,
			Date:       d,
			Qty:        r.Qty,
			Status:     models.WorkReportStatusSubmitted,
		})
	}

	if len(reports) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no reports"})
		return
	}

	if err := h.db.Create(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

type DeliveryInput struct {
	WorkItemID *uint   `json:"work_item_id"`
	Material   string  `json:"material" binding:"required"`
	Qty        float64 `json:"qty" binding:"required"`
	Date       string  `json:"date" binding:"required"`
}

func (h *Handler) CreateDelivery(c *gin.Context) {
	foremanID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	var obj models.Object
	if err := h.db.
		Where("id = ? AND foreman_user_id = ?", objectID, foremanID).
		First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}

	var in DeliveryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}

	delivery := models.MaterialDelivery{
		ObjectID:   obj.ID,
		WorkItemID: in.WorkItemID,
		ForemanID:  foremanID,
		Date:       d,
		Material:   in.Material,
		Qty:        in.Qty,
		Source:     "MANUAL",
	}

	if err := h.db.Create(&delivery).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

func (h *Handler) CreateDeliveryFromCV(c *gin.Context) {
	if h.cvProcessor == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cv processor is not configured"})
		return
	}

	foremanID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	var obj models.Object
	if err := h.db.
		Where("id = ? AND foreman_user_id = ?", objectID, foremanID).
		First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	workItemID, err := parseOptionalUint(c.PostForm("work_item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid work_item_id"})
		return
	}

	deliveryDate, err := parseOptionalDate(c.PostForm("date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}
	if deliveryDate == nil {
		now := time.Now()
		deliveryDate = &now
	}

	tempPath, err := h.saveMultipartToTemp(fileHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save uploaded file"})
		return
	}
	defer os.Remove(tempPath)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	cvResult, err := h.cvProcessor.ProcessFile(ctx, tempPath, fileHeader.Filename)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	finalPath, err := h.moveToPermanentStorage(obj.ID, fileHeader.Filename, tempPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to move file to storage"})
		return
	}

	delivery := models.MaterialDelivery{
		ObjectID:       obj.ID,
		WorkItemID:     workItemID,
		ForemanID:      foremanID,
		Date:           *deliveryDate,
		Material:       firstNonEmpty(ptrStringValue(cvResult.Extraction.MaterialName), fileHeader.Filename),
		Qty:            float64(ptrIntValue(cvResult.Extraction.Quantity)),
		Unit:           ptrStringValue(cvResult.Extraction.Unit),
		DocumentNumber: ptrStringValue(cvResult.Extraction.DocumentNumber),
		Source:         "CV",
		CVConfidence:   cvResult.Extraction.Confidence,
	}

	document := models.MaterialDocument{
		DocumentType:     models.MaterialDocumentTypeTTN,
		StoragePath:      finalPath,
		OriginalFileName: fileHeader.Filename,
		MimeType:         fileHeader.Header.Get("Content-Type"),
		CVStatus:         models.CVProcessingStatusDone,
		CVPayloadJSON:    string(cvResult.RawJSON),
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&delivery).Error; err != nil {
			return err
		}
		document.DeliveryID = delivery.ID
		if err := tx.Create(&document).Error; err != nil {
			return err
		}
		delivery.Documents = []models.MaterialDocument{document}
		return nil
	}); err != nil {
		_ = os.Remove(finalPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	var cvPayload map[string]any
	if err := json.Unmarshal(cvResult.RawJSON, &cvPayload); err != nil {
		cvPayload = map[string]any{"raw": string(cvResult.RawJSON)}
	}

	c.JSON(http.StatusCreated, gin.H{
		"delivery":   delivery,
		"document":   document,
		"cv_payload": cvPayload,
	})
}

func (h *Handler) saveMultipartToTemp(fileHeader *multipart.FileHeader) (string, error) {
	tempDir := filepath.Join(h.storageRoot, "tmp")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return "", err
	}

	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	tempPath := filepath.Join(tempDir, fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizeFilename(fileHeader.Filename)))
	dst, err := os.Create(tempPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return tempPath, nil
}

func (h *Handler) moveToPermanentStorage(objectID uint, originalName, tempPath string) (string, error) {
	targetDir := filepath.Join(h.storageRoot, "material-documents", fmt.Sprintf("object-%d", objectID))
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", err
	}

	finalPath := filepath.Join(targetDir, fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizeFilename(originalName)))
	if err := os.Rename(tempPath, finalPath); err != nil {
		return "", err
	}
	return finalPath, nil
}

func parseOptionalUint(value string) (*uint, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return nil, err
	}
	result := uint(parsed)
	return &result, nil
}

func parseOptionalDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func sanitizeFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "_")
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_")
	name = replacer.Replace(name)
	if name == "" {
		return "upload.bin"
	}
	return name
}

func ptrStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func ptrIntValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func HandlerForTest(db *gorm.DB) *Handler {
	return &Handler{
		db:          db,
		storageRoot: filepath.Join(os.TempDir(), "diploma-tests"),
	}
}

func HandlerForTestWithDeps(db *gorm.DB, cvProcessor cv.FileProcessor, storageRoot string) *Handler {
	return &Handler{
		db:          db,
		cvProcessor: cvProcessor,
		storageRoot: storageRoot,
	}
}
