package foreman_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/internal/auth"
	"github.com/Loc-Leonard/Diploma/internal/cv"
	"github.com/Loc-Leonard/Diploma/internal/foreman"
	"github.com/Loc-Leonard/Diploma/internal/models"
)

type mockCVProcessor struct {
	result *cv.FileProcessResult
	err    error
}

func (m mockCVProcessor) ProcessFile(_ context.Context, _ string, _ string) (*cv.FileProcessResult, error) {
	return m.result, m.err
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Fatalf("TEST_DB_DSN is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres db: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Object{},
		&models.WorkItem{},
		&models.MaterialDelivery{},
		&models.MaterialDocument{},
	); err != nil {
		t.Fatalf("migrate db: %v", err)
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

func setupForemanRouter(t *testing.T, db *gorm.DB, userID uint, processor cv.FileProcessor, storageRoot string) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := foreman.HandlerForTestWithDeps(db, processor, storageRoot)

	gr := r.Group("/foreman")
	gr.Use(fakeAuth(userID, string(models.RoleForeman)), auth.ForemanOnly())
	{
		gr.POST("/objects/:id/deliveries/cv-upload", h.CreateDeliveryFromCV)
		gr.GET("/objects/:id", h.ObjectDetail)
	}

	return r
}

func TestCreateDeliveryFromCV_CreatesDeliveryAndMaterialDocument(t *testing.T) {
	db := newTestDB(t)
	storageRoot := t.TempDir()

	foremanUser := models.User{
		FullName:           "Foreman Test",
		Role:               models.RoleForeman,
		MustChangePassword: false,
	}
	if err := db.Create(&foremanUser).Error; err != nil {
		t.Fatalf("seed foreman: %v", err)
	}

	obj := models.Object{
		Name:          "Object A",
		City:          "Moscow",
		Address:       "Test street",
		Status:        models.ObjectStatusActive,
		ForemanUserID: foremanUser.ID,
	}
	if err := db.Create(&obj).Error; err != nil {
		t.Fatalf("seed object: %v", err)
	}

	workItem := models.WorkItem{
		ObjectID: obj.ID,
		Name:     "Bordure",
		Unit:     "шт",
		PlanQty:  100,
	}
	if err := db.Create(&workItem).Error; err != nil {
		t.Fatalf("seed work item: %v", err)
	}

	materialName := "Бортовой камень 100x30"
	unit := "шт"
	documentNumber := "12345678"
	quantity := 25
	rawJSON := json.RawMessage(`{"extraction":{"material_name":"Бортовой камень 100x30","quantity":25}}`)

	processor := mockCVProcessor{
		result: &cv.FileProcessResult{
			SourcePath:    "temp/file.pdf",
			PredictedType: "ttn",
			Extraction: cv.ExtractionEntry{
				MaterialName:   &materialName,
				Quantity:       &quantity,
				Unit:           &unit,
				DocumentNumber: &documentNumber,
				Confidence:     0.93,
			},
			RawJSON: rawJSON,
		},
	}

	router := setupForemanRouter(t, db, foremanUser.ID, processor, storageRoot)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("file", "ttn.pdf")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("fake-pdf-content")); err != nil {
		t.Fatalf("write file content: %v", err)
	}
	if err := writer.WriteField("work_item_id", fmt.Sprintf("%d", workItem.ID)); err != nil {
		t.Fatalf("write work item field: %v", err)
	}
	if err := writer.WriteField("date", "2026-05-05"); err != nil {
		t.Fatalf("write date field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/foreman/objects/%d/deliveries/cv-upload", obj.ID), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var delivery models.MaterialDelivery
	if err := db.Preload("Documents").First(&delivery).Error; err != nil {
		t.Fatalf("load delivery: %v", err)
	}

	if delivery.ObjectID != obj.ID || delivery.ForemanID != foremanUser.ID {
		t.Fatalf("unexpected delivery relation: %+v", delivery)
	}
	if delivery.Material != materialName || delivery.Qty != float64(quantity) || delivery.Unit != unit {
		t.Fatalf("unexpected delivery payload: %+v", delivery)
	}
	if len(delivery.Documents) != 1 {
		t.Fatalf("expected 1 document, got %d", len(delivery.Documents))
	}

	document := delivery.Documents[0]
	if document.DocumentType != models.MaterialDocumentTypeTTN {
		t.Fatalf("unexpected document type: %s", document.DocumentType)
	}
	if document.CVStatus != models.CVProcessingStatusDone {
		t.Fatalf("unexpected cv status: %s", document.CVStatus)
	}
	if document.CVPayloadJSON != string(rawJSON) {
		t.Fatalf("unexpected raw json: %s", document.CVPayloadJSON)
	}
	if _, err := os.Stat(document.StoragePath); err != nil {
		t.Fatalf("stored file missing: %v", err)
	}
	if filepath.Dir(document.StoragePath) == storageRoot {
		t.Fatalf("expected file to be placed in nested storage directory, got %s", document.StoragePath)
	}

	var response struct {
		Delivery models.MaterialDelivery `json:"delivery"`
		Document models.MaterialDocument `json:"document"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Delivery.Material != materialName {
		t.Fatalf("unexpected response delivery: %+v", response.Delivery)
	}
	if response.Document.OriginalFileName != "ttn.pdf" {
		t.Fatalf("unexpected response document: %+v", response.Document)
	}
}

func TestObjectDetail_ReturnsDeliveriesWithDocuments(t *testing.T) {
	db := newTestDB(t)

	foremanUser := models.User{
		FullName:           "Foreman Test",
		Role:               models.RoleForeman,
		MustChangePassword: false,
	}
	if err := db.Create(&foremanUser).Error; err != nil {
		t.Fatalf("seed foreman: %v", err)
	}

	obj := models.Object{
		Name:          "Object A",
		City:          "Moscow",
		Address:       "Test street",
		Status:        models.ObjectStatusActive,
		ForemanUserID: foremanUser.ID,
	}
	if err := db.Create(&obj).Error; err != nil {
		t.Fatalf("seed object: %v", err)
	}

	delivery := models.MaterialDelivery{
		ObjectID:       obj.ID,
		ForemanID:      foremanUser.ID,
		Date:           time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
		Material:       "Material",
		Qty:            10,
		Source:         "CV",
		DocumentNumber: "12345678",
	}
	if err := db.Create(&delivery).Error; err != nil {
		t.Fatalf("seed delivery: %v", err)
	}

	document := models.MaterialDocument{
		DeliveryID:       delivery.ID,
		DocumentType:     models.MaterialDocumentTypeTTN,
		StoragePath:      "storage/file.pdf",
		OriginalFileName: "file.pdf",
		MimeType:         "application/pdf",
		CVStatus:         models.CVProcessingStatusDone,
		CVPayloadJSON:    "{}",
	}
	if err := db.Create(&document).Error; err != nil {
		t.Fatalf("seed document: %v", err)
	}

	router := setupForemanRouter(t, db, foremanUser.ID, nil, t.TempDir())

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/foreman/objects/%d", obj.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response struct {
		Deliveries []models.MaterialDelivery `json:"deliveries"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(response.Deliveries) != 1 || len(response.Deliveries[0].Documents) != 1 {
		t.Fatalf("expected nested documents in response, got %+v", response.Deliveries)
	}
}
