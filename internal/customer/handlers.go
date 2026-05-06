package customer

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

		gr.GET("objects/:id/work-items", h.ListWorkItems)
		gr.POST("objects/:id/work-items", h.CreateWorkItem)
		gr.PUT("objects/:id/work-items/:wid", h.UpdateWorkItem)
		gr.DELETE("objects/:id/work-items/:wid", h.DeleteWorkItem)
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

	Progress float64 `json:"progress"`

	Foreman *struct {
		ID       uint   `json:"id"`
		FullName string `json:"full_name"`
	} `json:"foreman,omitempty"`

	PlannedStartDate *time.Time `json:"planned_start_date"`
	PlannedEndDate   *time.Time `json:"planned_end_date"`

	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`

	ActivationRejectReason string `json:"activation_reject_reason"`
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

type ActivateObjectRequest struct {
	ChecklistJSON string  `json:"checklist_json" binding:"required"`
	ActFilePath   *string `json:"act_file_path"`
}

// POST /customer/objects/:id/activate
func (h *Handler) ActivateObject(c *gin.Context) {
	customerID := auth.UserIDFromContext(c)
	if customerID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")

	var obj models.Object
	if err := h.db.
		Where("id = ? AND customer_control_user_id = ?", id, customerID).
		First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}

	if obj.Status != models.ObjectStatusPlanned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "object is not in PLANNED status"})
		return
	}

	var req ActivateObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "sent_for_inspector_confirmation"})
}

// for tests
func HandlerForTest(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// GET /customer/objects/:id
func (h *Handler) GetObject(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	id := c.Param("id")

	obj, err := objectcore.LoadObjectForUser(h.db, id, userID, string(models.RoleCustomer))
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

func (h Handler) ListWorkItems(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	//Проверка, что объект существует и принадлежит этому заказчику

	var obj models.Object
	if err := h.db.Where("id = ? AND customer_control_user_id = ?", objectID, userID).
		First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}
	var items []models.WorkItem
	h.db.Where("object_id = ?", obj.ID).Order("sort_order ASC, id ASC").Find(&items)
	c.JSON(http.StatusOK, items)
}

type WorkItemInput struct {
	Name             string     `json:"name" binding:"required"`
	Description      string     `json:"description"`
	Unit             string     `json:"unit"`
	PlanQty          float64    `json:"plan_qty"`
	PlannedStartDate *time.Time `json:"planned_start_date"`
	PlannedEndDate   *time.Time `json:"planned_end_date"`
	SortOrder        int        `json:"sort_order"`
	DependsOnID      *uint      `json:"depends_on_id"`
}

func (h Handler) CreateWorkItem(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")

	var obj models.Object
	if err := h.db.Where("id = ? AND customer_control_user_id = ?", objectID, userID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}

	var in WorkItemInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h Handler) UpdateWorkItem(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")
	wid := c.Param("wid")

	//Проверка доступа к объекту
	var obj models.Object
	if err := h.db.Where("id = ? AND customer_control_user_id = ?", objectID, userID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}

	var item models.WorkItem
	if err := h.db.Where("id = ? AND object_id = ?", wid, obj.ID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work item not found"})
		return
	}

	var in WorkItemInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	h.db.Save(&item)
	c.JSON(http.StatusOK, item)

}

func (h Handler) DeleteWorkItem(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	objectID := c.Param("id")
	wid := c.Param("wid")

	var obj models.Object
	if err := h.db.Where("id = ? AND customer_control_user_id = ?", objectID, userID).First(&obj).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
		return
	}

	h.db.Where("id = ? AND object_id = ?", wid, obj.ID).Delete(&models.WorkItem{})
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
