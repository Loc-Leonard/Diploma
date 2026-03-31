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

		gr.GET("/foremen-list", h.ForemenList)
		gr.GET("/inspectors-list", h.InspectorsList)

		gr.POST("/objects", h.CreateObject)
	}
}

type SimpleUserDTO struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
	City     string `json:"city,omitempty"`
}

func (h *Handler) ForemenList(c *gin.Context) {
	var users []models.User
	if err := h.db.
		Where("role = ?", models.RoleForeman).
		Order("full_name ASC").
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	resp := make([]SimpleUserDTO, 0, len(users))
	for _, u := range users {
		resp = append(resp, SimpleUserDTO{
			ID:       u.ID,
			FullName: u.FullName,
			//City:,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	resp := make([]SimpleUserDTO, 0, len(users))
	for _, u := range users {
		resp = append(resp, SimpleUserDTO{
			ID:       u.ID,
			FullName: u.FullName,
		})
	}

	c.JSON(http.StatusOK, resp)
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

// DTO for Create object
type CreateObjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Address     string `json:"address" binding:"required"`
	City        string `json:"city" binding:"required"`
	Description string `json:"description"`

	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`

	PlannedStartDate *time.Time `json:"planned_start_date"`
	PlannedEndDate   *time.Time `json:"planned_end_date"`

	ForemanUserID   uint `json:"foreman_user_id" binding:"required"`
	InspectorUserID uint `json:"inspector_user_id" binding:"required"`
}

func (h *Handler) CreateObject(c *gin.Context) {
	customerID := auth.UserIDFromContext(c)

	var req CreateObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var inspector models.User
	if err := h.db.
		Where("id = ? AND role = ?", req.InspectorUserID, models.RoleInspector).
		First(&inspector).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inspector"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": obj.ID})
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
