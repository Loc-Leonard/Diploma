package inspector

import (
	//"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/internal/auth"
	"github.com/Loc-Leonard/Diploma/internal/models"
	//"github.com/Loc-Leonard/Diploma/internal/objectcore"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	h := &Handler{db: db}

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
	}
}

type Handler struct {
	db *gorm.DB
}

type InspectorDashboardCheck struct {
	ID         uint      `json:"id"`
	ObjectID   uint      `json:"object_id"`
	ObjectName string    `json:"object_name"`
	City       string    `json:"city"`
	Address    string    `json:"address"`
	Status     string    `json:"status"`
	PlannedAt  time.Time `json:"planned_at"`
	IssuesOpen int       `json:"issues_open"`
}

type InspectorDashboardObject struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	City          string `json:"city"`
	Address       string `json:"address"`
	ForemanName   string `json:"foreman_name"`
	ActiveChecks  int    `json:"active_checks"`
	OverdueChecks int    `json:"overdue_checks"`
	OpenIssues    int    `json:"open_issues"`
}

type InspectorObjectListItem struct {
	ID               uint                `json:"id"`
	Name             string              `json:"name"`
	City             string              `json:"city"`
	Address          string              `json:"address"`
	Status           models.ObjectStatus `json:"status"`
	ForemanName      string              `json:"foreman_name"`
	PlannedStartDate *time.Time          `json:"planned_start_date"`
	HasPendingAction bool                `json:"has_pending_action"`
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
			o.planned_start_date
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
				o.planned_start_date
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

	type PersonDTO struct {
		ID       uint   `json:"id"`
		FullName string `json:"full_name"`
	}

	type ObjectDTO struct {
		ID                     uint                `json:"id"`
		Name                   string              `json:"name"`
		City                   string              `json:"city"`
		Address                string              `json:"address"`
		Description            string              `json:"description"`
		Status                 models.ObjectStatus `json:"status"`
		Lat                    float64             `json:"lat"`
		Lng                    float64             `json:"lng"`
		PlannedStartDate       *time.Time          `json:"planned_start_date"`
		PlannedEndDate         *time.Time          `json:"planned_end_date"`
		ActualStartDate        *time.Time          `json:"actual_start_date"`
		Customer               *PersonDTO          `json:"customer,omitempty"`
		Foreman                *PersonDTO          `json:"foreman,omitempty"`
		Inspector              *PersonDTO          `json:"inspector,omitempty"`
		InitChecklistJSON      string              `json:"init_checklist_json"`
		InitActFilePath        string              `json:"init_act_file_path"`
		ActivationRejectReason string              `json:"activation_reject_reason"`
	}

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

type ActivationDecisionRequest struct {
	Decision        string `json:"decision" binding:"required"`
	RejectionReason string `json:"rejection_reason"`
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

// for tests
func HandlerForTest(db *gorm.DB) *Handler {
	return &Handler{db: db}
}
