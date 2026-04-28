package foreman

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/internal/auth"
	"github.com/Loc-Leonard/Diploma/internal/models"
	"github.com/Loc-Leonard/Diploma/internal/objectcore"
)

type Handler struct {
	db *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	h := &Handler{db: db}

	gr := r.Group("/foreman")
	gr.Use(
		auth.AuthMiddleware(),
		auth.ForemanOnly(),
		auth.MustChangePasswordMiddleware(db))

	{
		gr.GET("/objects", h.ListObjects)
		gr.GET("/objects/:id", h.ObjectDetail)
		gr.POST("/objects/:id/work-reports", h.CreateWorkReports)
		gr.POST("/objects/:id/deliveries", h.CreateDelivery)
	}
}

// DTO для списка объектов прораба
type ForemanObjectDTO struct {
	ID      uint                `json:"id"`
	Name    string              `json:"name"`
	City    string              `json:"city"`
	Address string              `json:"address"`
	Status  models.ObjectStatus `json:"status"`
}

// GET /foreman/objects
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

// DTO деталки объекта
type ForemanObjectDetailDTO struct {
	Object    ForemanObjectDTO  `json:"object"`
	WorkItems []models.WorkItem `json:"work_items"`
}

// GET /foreman/objects/:id
func (h Handler) ObjectDetail(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	id := c.Param("id")

	obj, err := objectcore.LoadObjectForUser(h.db, id, userID, string(models.RoleForeman))
	if errors.Is(err, objectcore.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}
	if errors.Is(err, objectcore.ErrForbidden) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	core := objectcore.BuildObjectCoreDTO(h.db, obj)

	var items []models.WorkItem
	h.db.Where("object_id = ?", obj.ID).Order("id ASC").Find(&items)

	c.JSON(http.StatusOK, gin.H{
		"object":     core,
		"work_items": items,
	})
}

// POST /foreman/objects/:id/work-reports
type WorkReportInput struct {
	WorkItemID uint    `json:"work_item_id"`
	Qty        float64 `json:"qty"`
	Date       string  `json:"date"` // "2006-01-02"
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

// POST /foreman/objects/:id/deliveries
type DeliveryInput struct {
	WorkItemID *uint   `json:"work_item_id"`
	Material   string  `json:"material" binding:"required"`
	Qty        float64 `json:"qty" binding:"required"`
	Date       string  `json:"date" binding:"required"` // "2006-01-02"
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
