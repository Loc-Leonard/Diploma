package inspector

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/internal/auth"
	"github.com/Loc-Leonard/Diploma/internal/models"
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
	}
}

type Handler struct {
	db *gorm.DB
}

// DTO под дашборд проверок
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

// DTO под дашборд объектов инспектора
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

// for tests
func HandlerForTest(db *gorm.DB) *Handler {
	return &Handler{db: db}
}
