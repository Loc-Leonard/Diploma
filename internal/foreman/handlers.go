package foreman

import (
	"errors"
	"math"
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

	detail := objectcore.BuildObjectDetailDTO(h.db, obj)
	c.JSON(http.StatusOK, detail)
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

	// 1. Проверяем доступ к объекту
	var obj models.Object
	if err := h.db.
		Where("id = ? AND foreman_user_id = ?", objectID, foremanID).
		First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}

	// 2. Читаем тело запроса
	var req WorkReportsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// 3. Собираем валидные отчёты
	reports := make([]models.WorkReport, 0, len(req.Reports))
	for _, r := range req.Reports {
		if r.Qty <= 0 {
			continue
		}
		d, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date: " + r.Date})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid reports"})
		return
	}

	// 4. Сохраняем отчёты
	if err := h.db.Create(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	// 5. Пересчитываем прогресс — только после сохранения отчётов
	recalcProgress(h.db, obj.ID)

	// 6. Отвечаем клиенту
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

// recalcProgress пересчитывает прогресс каждого этапа и общий прогресс объекта
// на основе суммы фактических qty из отчётов относительно планового qty
func recalcProgress(db *gorm.DB, objectID uint) error {
	var items []models.WorkItem
	db.Where("object_id = ?", objectID).Find(&items)
	if len(items) == 0 {
		return nil
	}

	totalPlan := 0.0
	totalFact := 0.0

	for _, item := range items {
		var factQty float64
		db.Model(&models.WorkReport{}).
			Where("work_item_id = ?", item.ID).
			Select("COALESCE(SUM(qty), 0)").
			Scan(&factQty)

		itemProgress := 0.0
		if item.PlanQty > 0 {
			itemProgress = math.Min(factQty/item.PlanQty*100, 100)
		}

		db.Model(&models.WorkItem{}).
			Where("id = ?", item.ID).
			Update("progress", itemProgress)

		totalPlan += item.PlanQty
		totalFact += factQty
	}

	objectProgress := 0.0
	if totalPlan > 0 {
		objectProgress = math.Min(totalFact/totalPlan*100, 100)
	}

	return db.Model(&models.Object{}).
		Where("id = ?", objectID).
		Update("progress", objectProgress).Error
}
