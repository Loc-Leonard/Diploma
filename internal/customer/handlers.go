package customer

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/internal/auth"
	"github.com/Loc-Leonard/Diploma/internal/models"
)

type Handler struct {
	db *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	h := &Handler{db: db}

	gr := r.Group("/customer")
	gr.Use(auth.AuthMiddleware(), auth.MustChangePasswordMiddleware(db), auth.CustomerOnly())
	{
		gr.GET("/dashboard/objects", h.DashboardObjects)
		gr.GET("/dashboard/foremen", h.DashboardForemen)
	}
}

type DashboardObjectDTO struct {
	ID      uint                `json:"id"`
	Name    string              `json:"name"`
	City    string              `json:"city"`
	Address string              `json:"address"`
	Status  models.ObjectStatus `json:"status"`

	Progress int `json:"progress"`

	Foreman *struct {
		ID       uint   `json:"id"`
		FullName string `json:"full_name"`
	} `json:"foreman,omitempty"`

	PlannedStartDate *time.Time `json:"planned_start_date"`
	PlannedEndDate   *time.Time `json:"planned_end_date"`

	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// DTO прораба
type DashboardForemanDTO struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
	City     string `json:"city"`

	CurrentObject *struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	} `json:"current_object,omitempty"`
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
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
			ID:               o.ID,
			Name:             o.Name,
			City:             o.City,
			Address:          o.Address,
			Status:           o.Status,
			Progress:         0, // потом посчитаем
			PlannedStartDate: o.PlannedStartDate,
			PlannedEndDate:   o.PlannedEndDate,
			Lat:              o.Lat,
			Lng:              o.Lng,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
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

// for tests
func HandlerForTest(db *gorm.DB) *Handler {
	return &Handler{db: db}
}
