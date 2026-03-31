package admin_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/internal/admin"
	"github.com/Loc-Leonard/Diploma/internal/auth"
	"github.com/Loc-Leonard/Diploma/internal/models"
)

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
	return db
}
func fakeAuth(userID uint, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

func setupAdminRouter(t *testing.T, db *gorm.DB, userID uint, role string) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := admin.HandlerForTest(db) // см. ниже комментарий

	gr := r.Group("/admin")
	gr.Use(
		fakeAuth(userID, role),
		auth.AdminOnly(),
	)
	{
		gr.POST("/users", h.CreateUser)
		gr.GET("/users", h.ListUsers)
	}

	return r
}

func TestAdmin_ListUsers_Empty(t *testing.T) {
	db := newTestDB(t)

	adminUser := models.User{
		FullName: "Admin",
		Role:     models.RoleAdmin,
	}
	if err := db.Create(&adminUser).Error; err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	r := setupAdminRouter(t, db, adminUser.ID, string(models.RoleAdmin))

	req, _ := http.NewRequest(http.MethodGet, "/admin/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp []admin.UserListItem
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 user (admin), got %d", len(resp))
	}
}

func TestAdmin_CreateUser_And_ListUsers(t *testing.T) {
	db := newTestDB(t)

	adminUser := models.User{
		FullName: "Admin",
		Role:     models.RoleAdmin,
	}
	if err := db.Create(&adminUser).Error; err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	r := setupAdminRouter(t, db, adminUser.ID, string(models.RoleAdmin))

	body := admin.CreateUserRequest{
		FullName: "new user",
		Email:    strPtr("new@example.com"),
		Phone:    strPtr("+79001231234"),
		Role:     models.RoleCustomer,
	}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/admin/users", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	if created["full_name"] != body.FullName {
		t.Fatalf("expected full_name=%q, got %v", body.FullName, created["full_name"])
	}
	if created["role"] != string(body.Role) {
		t.Fatalf("expected role=%q, got %v", body.Role, created["role"])
	}
	if created["temp_password"] == "" || created["temp_password"] == nil {
		t.Fatalf("expected temp_password in response")
	}

	//check ListUsers
	req2, _ := http.NewRequest(http.MethodGet, "/admin/users", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusOK, w2.Code, w2.Body.String())
	}

	var list []admin.UserListItem
	if err := json.Unmarshal(w2.Body.Bytes(), &list); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}

	found := false
	for _, u := range list {
		if u.FullName == body.FullName && u.Role == body.Role {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created user noot found in list, list=%+v", list)
	}
}

func TestAdmin_CreateUser_ValidationError(t *testing.T) {
	db := newTestDB(t)

	adminUser := models.User{
		FullName: "Admin",
		Role:     models.RoleAdmin,
	}
	if err := db.Create(&adminUser).Error; err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	r := setupAdminRouter(t, db, adminUser.ID, string(models.RoleAdmin))

	req, _ := http.NewRequest(http.MethodPost, "/admin/users", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expaected %d, got %d, body: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestAdmin_AccessDenied_ForNonAdmin(t *testing.T) {
	db := newTestDB(t)

	user := models.User{
		FullName: "non Admin",
		Role:     models.RoleCustomer,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	r := setupAdminRouter(t, db, user.ID, string(models.RoleCustomer))
	req, _ := http.NewRequest(http.MethodGet, "/admin/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusForbidden, w.Code, w.Body.String())
	}
}

func strPtr(s string) *string {
	return &s
}
