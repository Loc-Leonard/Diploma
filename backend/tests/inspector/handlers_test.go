package inspector_test

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

	"github.com/Loc-Leonard/Diploma/backend/internal/auth"
	"github.com/Loc-Leonard/Diploma/backend/internal/inspector"
	"github.com/Loc-Leonard/Diploma/backend/internal/models"
)

// ==== DTO для ответов (копия публичных) ====

type inspectorDashboardCheck struct {
	ID         uint      `json:"id"`
	ObjectID   uint      `json:"object_id"`
	ObjectName string    `json:"object_name"`
	City       string    `json:"city"`
	Address    string    `json:"address"`
	Status     string    `json:"status"`
	PlannedAt  time.Time `json:"planned_at"`
	IssuesOpen int       `json:"issues_open"`
}

type inspectorDashboardObject struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	City          string `json:"city"`
	Address       string `json:"address"`
	ForemanName   string `json:"foreman_name"`
	ActiveChecks  int    `json:"active_checks"`
	OverdueChecks int    `json:"overdue_checks"`
	OpenIssues    int    `json:"open_issues"`
}

// ==== DB и фейковый auth ====

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Fatalf("TEST_DB_DSN is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Object{}, &models.Inspection{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func fakeAuth(userID uint, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

func setupInspectorRouter(t *testing.T, db *gorm.DB, userID uint, role string) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := inspector.HandlerForTest(db)

	gr := r.Group("/inspector")
	gr.Use(
		fakeAuth(userID, role),
		auth.MustChangePasswordMiddleware(db),
		auth.InspectorOnly(),
	)
	{
		gr.GET("/dashboard/checks", h.DashboardChecks)
		gr.GET("/dashboard/objects", h.DashboardObjects)
	}

	return r
}

// ==== Тесты ====

func TestInspectorDashboardChecks_Basic(t *testing.T) {
	db := newTestDB(t)

	// инспектор
	insp := models.User{
		FullName: "Инспектор",
		Role:     models.RoleInspector,
	}
	if err := db.Create(&insp).Error; err != nil {
		t.Fatalf("seed inspector: %v", err)
	}

	// объект
	obj := models.Object{
		Name:    "Объект 1",
		City:    "Москва",
		Address: "Адрес",
	}
	if err := db.Create(&obj).Error; err != nil {
		t.Fatalf("seed object: %v", err)
	}

	now := time.Now().Truncate(time.Second)

	ins1 := models.Inspection{
		ObjectID:    obj.ID,
		InspectorID: insp.ID,
		Status:      models.InspectionStatusPlanned,
		PlannedAt:   now,
		IssuesOpen:  3,
	}
	if err := db.Create(&ins1).Error; err != nil {
		t.Fatalf("seed inspection: %v", err)
	}

	r := setupInspectorRouter(t, db, insp.ID, string(models.RoleInspector))

	req, _ := http.NewRequest(http.MethodGet, "/inspector/dashboard/checks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp []inspectorDashboardCheck
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 check, got %d", len(resp))
	}

	got := resp[0]
	if got.ObjectID != obj.ID ||
		got.ObjectName != obj.Name ||
		got.City != obj.City ||
		got.Address != obj.Address ||
		got.Status != string(models.InspectionStatusPlanned) ||
		got.IssuesOpen != 3 {
		t.Fatalf("unexpected check data: %+v", got)
	}
}

func TestInspectorDashboardChecks_Filters(t *testing.T) {
	db := newTestDB(t)

	insp := models.User{
		FullName: "Инспектор",
		Role:     models.RoleInspector,
	}
	if err := db.Create(&insp).Error; err != nil {
		t.Fatalf("seed inspector: %v", err)
	}

	obj1 := models.Object{Name: "МСК объект", City: "Москва", Address: "А"}
	obj2 := models.Object{Name: "СПб объект", City: "СПб", Address: "Б"}
	if err := db.Create(&obj1).Error; err != nil {
		t.Fatalf("seed obj1: %v", err)
	}
	if err := db.Create(&obj2).Error; err != nil {
		t.Fatalf("seed obj2: %v", err)
	}

	now := time.Now().Truncate(time.Second)

	ins1 := models.Inspection{
		ObjectID:    obj1.ID,
		InspectorID: insp.ID,
		Status:      models.InspectionStatusPlanned,
		PlannedAt:   now,
	}
	ins2 := models.Inspection{
		ObjectID:    obj2.ID,
		InspectorID: insp.ID,
		Status:      models.InspectionStatusOverdue,
		PlannedAt:   now,
	}
	if err := db.Create(&ins1).Error; err != nil {
		t.Fatalf("seed ins1: %v", err)
	}
	if err := db.Create(&ins2).Error; err != nil {
		t.Fatalf("seed ins2: %v", err)
	}

	r := setupInspectorRouter(t, db, insp.ID, string(models.RoleInspector))

	// фильтр по статусу + городу
	req, _ := http.NewRequest(
		http.MethodGet,
		"/inspector/dashboard/checks?status="+string(models.InspectionStatusPlanned)+"&city=Москва",
		nil,
	)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp []inspectorDashboardCheck
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 check after filter, got %d", len(resp))
	}
	if resp[0].ObjectID != obj1.ID {
		t.Fatalf("expected object %d, got %d", obj1.ID, resp[0].ObjectID)
	}
}

func TestInspectorDashboardObjects_Basic(t *testing.T) {
	db := newTestDB(t)

	insp := models.User{
		FullName: "Инспектор",
		Role:     models.RoleInspector,
	}
	if err := db.Create(&insp).Error; err != nil {
		t.Fatalf("seed inspector: %v", err)
	}

	foreman := models.User{
		FullName: "Прораб",
		Role:     models.RoleForeman,
	}
	if err := db.Create(&foreman).Error; err != nil {
		t.Fatalf("seed foreman: %v", err)
	}

	obj := models.Object{
		Name:          "Объект 1",
		City:          "Москва",
		Address:       "Адрес",
		ForemanUserID: foreman.ID,
	}
	if err := db.Create(&obj).Error; err != nil {
		t.Fatalf("seed object: %v", err)
	}

	now := time.Now().Truncate(time.Second)

	insPlanned := models.Inspection{
		ObjectID:    obj.ID,
		InspectorID: insp.ID,
		Status:      models.InspectionStatusPlanned,
		PlannedAt:   now,
		IssuesOpen:  2,
	}
	insOverdue := models.Inspection{
		ObjectID:    obj.ID,
		InspectorID: insp.ID,
		Status:      models.InspectionStatusOverdue,
		PlannedAt:   now,
		IssuesOpen:  1,
	}
	if err := db.Create(&insPlanned).Error; err != nil {
		t.Fatalf("seed planned: %v", err)
	}
	if err := db.Create(&insOverdue).Error; err != nil {
		t.Fatalf("seed overdue: %v", err)
	}

	r := setupInspectorRouter(t, db, insp.ID, string(models.RoleInspector))

	req, _ := http.NewRequest(http.MethodGet, "/inspector/dashboard/objects", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp []inspectorDashboardObject
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
		got.Address != obj.Address {
		t.Fatalf("unexpected object data: %+v", got)
	}
	if got.ForemanName != foreman.FullName {
		t.Fatalf("unexpected foreman_name: %q", got.ForemanName)
	}
	if got.ActiveChecks != 1 || got.OverdueChecks != 1 || got.OpenIssues != 3 {
		t.Fatalf("unexpected counters: %+v", got)
	}
}
