package customer_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/internal/auth"
	"github.com/Loc-Leonard/Diploma/internal/customer"
	"github.com/Loc-Leonard/Diploma/internal/models"
)

// используем реальные DTO, чтобы точно проверять JSON
type dashboardObjectDTO struct {
	ID       uint                `json:"id"`
	Name     string              `json:"name"`
	City     string              `json:"city"`
	Address  string              `json:"address"`
	Status   models.ObjectStatus `json:"status"`
	Progress int                 `json:"progress"`
	Foreman  *struct {
		ID       uint   `json:"id"`
		FullName string `json:"full_name"`
	} `json:"foreman,omitempty"`
	PlannedStartDate *time.Time `json:"planned_start_date"`
	PlannedEndDate   *time.Time `json:"planned_end_date"`
	Lat              float64    `json:"lat"`
	Lng              float64    `json:"lng"`
}

type dashboardForemanDTO struct {
	ID            uint   `json:"id"`
	FullName      string `json:"full_name"`
	City          string `json:"city"`
	CurrentObject *struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	} `json:"current_object,omitempty"`
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Fatalf("TEST_DB_DSN is not set")
	}
	t.Logf("TEST_DB_DSN=%s", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Object{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

// фэйковый auth: подставляем user_id и role так же, как настоящие middleware
func fakeAuth(userID uint, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

func setupCustomerRouter(t *testing.T, db *gorm.DB, userID uint, role string) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := customer.HandlerForTest(db)

	gr := r.Group("/customer")
	// подменяем реальные AuthMiddleware(), MustChangePasswordMiddleware, CustomerOnly
	gr.Use(
		fakeAuth(userID, role),
		// MustChangePasswordMiddleware использует DB + auth.UserIDFromContext
		auth.MustChangePasswordMiddleware(db),
		auth.CustomerOnly(),
	)
	{
		gr.GET("/dashboard/objects", h.DashboardObjects)
		gr.GET("/dashboard/foremen", h.DashboardForemen)
	}

	return r
}

func TestDashboardObjects_Basic(t *testing.T) {
	db := newTestDB(t)

	customerID := uint(10)
	foremanID := uint(20)

	// прораб
	foreman := models.User{
		ID:       foremanID,
		FullName: "Иванов Иван",
		Role:     models.RoleForeman,
	}
	if err := db.Create(&foreman).Error; err != nil {
		t.Fatalf("seed foreman: %v", err)
	}

	// заказчик
	customerUser := models.User{
		ID:       customerID,
		FullName: "Заказчик",
		Role:     models.RoleCustomer,
	}
	if err := db.Create(&customerUser).Error; err != nil {
		t.Fatalf("seed customer: %v", err)
	}

	now := time.Now().Truncate(time.Second)

	obj := models.Object{
		Name:                  "Объект 1",
		Address:               "Москва, Тверская 1",
		City:                  "Москва",
		Description:           "описание",
		Status:                models.ObjectStatusActive,
		Lat:                   55.75,
		Lng:                   37.61,
		CustomerControlUserID: customerID,
		ForemanUserID:         foremanID,
		PlannedStartDate:      &now,
		PlannedEndDate:        &now,
	}
	if err := db.Create(&obj).Error; err != nil {
		t.Fatalf("seed object: %v", err)
	}

	r := setupCustomerRouter(t, db, customerID, string(models.RoleCustomer))

	req, _ := http.NewRequest(http.MethodGet, "/customer/dashboard/objects", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp []dashboardObjectDTO
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 object, got %d", len(resp))
	}
	got := resp[0]
	if got.ID != obj.ID ||
		got.Name != obj.Name ||
		got.City != obj.City ||
		got.Address != obj.Address ||
		got.Status != obj.Status ||
		got.Lat != obj.Lat ||
		got.Lng != obj.Lng {
		t.Fatalf("unexpected object data: %+v", got)
	}
	if got.Foreman == nil || got.Foreman.ID != foremanID || got.Foreman.FullName != foreman.FullName {
		t.Fatalf("unexpected foreman: %+v", got.Foreman)
	}
}

func TestDashboardObjects_Filters(t *testing.T) {
	db := newTestDB(t)
	customerID := uint(10)

	customerUser := models.User{
		ID:       customerID,
		FullName: "Заказчик",
		Role:     models.RoleCustomer,
	}
	if err := db.Create(&customerUser).Error; err != nil {
		t.Fatalf("seed customer: %v", err)
	}

	obj1 := models.Object{
		Name:                  "Объект Москва активный",
		City:                  "Москва",
		Status:                models.ObjectStatusActive,
		CustomerControlUserID: customerID,
	}
	obj2 := models.Object{
		Name:                  "Объект СПб завершённый",
		City:                  "СПб",
		Status:                models.ObjectStatusFinished,
		CustomerControlUserID: customerID,
	}
	if err := db.Create(&obj1).Error; err != nil {
		t.Fatalf("seed obj1: %v", err)
	}
	if err := db.Create(&obj2).Error; err != nil {
		t.Fatalf("seed obj2: %v", err)
	}

	r := setupCustomerRouter(t, db, customerID, string(models.RoleCustomer))

	req, _ := http.NewRequest(http.MethodGet,
		"/customer/dashboard/objects?city=Москва&status="+string(models.ObjectStatusActive),
		nil,
	)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp []dashboardObjectDTO
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 object after filter, got %d", len(resp))
	}
	if resp[0].Name != obj1.Name {
		t.Fatalf("expected %q, got %q", obj1.Name, resp[0].Name)
	}
}

func TestDashboardForemen_Basic(t *testing.T) {
	db := newTestDB(t)

	customerID := uint(10)
	foremanID := uint(20)

	customerUser := models.User{
		ID:       customerID,
		FullName: "Заказчик",
		Role:     models.RoleCustomer,
	}
	if err := db.Create(&customerUser).Error; err != nil {
		t.Fatalf("seed customer: %v", err)
	}

	foreman := models.User{
		ID:       foremanID,
		FullName: "Иванов Иван",
		Role:     models.RoleForeman,
	}
	if err := db.Create(&foreman).Error; err != nil {
		t.Fatalf("seed foreman: %v", err)
	}

	obj := models.Object{
		Name:                  "Объект 1",
		City:                  "Москва",
		CustomerControlUserID: customerID,
		ForemanUserID:         foremanID,
	}
	if err := db.Create(&obj).Error; err != nil {
		t.Fatalf("seed object: %v", err)
	}

	r := setupCustomerRouter(t, db, customerID, string(models.RoleCustomer))

	req, _ := http.NewRequest(http.MethodGet, "/customer/dashboard/foremen", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp []dashboardForemanDTO
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 foreman, got %d", len(resp))
	}
	got := resp[0]
	if got.ID != foremanID || got.FullName != foreman.FullName || got.City != obj.City {
		t.Fatalf("unexpected foreman data: %+v", got)
	}
	if got.CurrentObject == nil || got.CurrentObject.ID != obj.ID || got.CurrentObject.Name != obj.Name {
		t.Fatalf("unexpected current object: %+v", got.CurrentObject)
	}
}
