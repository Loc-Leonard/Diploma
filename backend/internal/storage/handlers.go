package storage

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Loc-Leonard/Diploma/backend/internal/auth"
	"github.com/Loc-Leonard/Diploma/backend/internal/models"
)

type Handler struct {
	db          *gorm.DB
	storageRoot string
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB, storageRoot string) {
	h := &Handler{
		db:          db,
		storageRoot: storageRoot,
	}

	// Endpoint для скачивания файлов - требует авторизации
	gr := r.Group("/api/storage")
	gr.Use(auth.AuthMiddleware())
	{
		gr.GET("/download/:documentId", h.DownloadFile)
	}
}

// DownloadFile - скачивание файла по ID документа
// GET /api/storage/download/:documentId
func (h *Handler) DownloadFile(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	documentID := c.Param("documentId")

	// Получаем документ
	var doc models.MaterialDocument
	if err := h.db.First(&doc, documentID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "document not found"})
		return
	}

	// Проверяем доступ пользователя к документу
	hasAccess := h.checkAccess(userID, doc)

	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "access denied"})
		return
	}

	// Проверяем существование файла
	fileInfo, err := os.Stat(doc.StoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "file not found on server"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "cannot access file"})
		return
	}

	// Проверяем, что это файл, а не директория
	if fileInfo.IsDir() {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "not a file"})
		return
	}

	// Открываем файл
	file, err := os.Open(doc.StoragePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "cannot open file"})
		return
	}
	defer file.Close()

	// Определяем MIME тип
	contentType := doc.MimeType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Устанавливаем заголовки для скачивания
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=\""+doc.OriginalFileName+"\"")
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "no-cache")

	// Отправляем файл
	c.File(doc.StoragePath)
}

// checkAccess - проверка доступа пользователя к документу
func (h *Handler) checkAccess(userID uint, doc models.MaterialDocument) bool {
	// 1. Блок проверки для замечаний
	if doc.IssueID != nil {
		var issue models.Issue
		if err := h.db.First(&issue, *doc.IssueID).Error; err == nil {
			var obj models.Object
			if err := h.db.First(&obj, issue.ObjectID).Error; err == nil {
				if obj.CustomerControlUserID == userID || obj.ForemanUserID == userID || obj.InspectorUserID == userID {
					return true
				}
			}
		}
	}
	// 2. Проверяем через UploadedBy
	if doc.UploadedBy != nil && *doc.UploadedBy == userID {
		return true
	}

	// 3. Проверяем через Delivery (для Foreman)
	if doc.DeliveryID != nil {
		var delivery models.MaterialDelivery
		if err := h.db.First(&delivery, *doc.DeliveryID).Error; err == nil {
			if delivery.ForemanID == userID {
				return true
			}
			// Проверяем доступ через объект
			var obj models.Object
			if err := h.db.First(&obj, delivery.ObjectID).Error; err == nil {
				if obj.CustomerControlUserID == userID ||
					obj.ForemanUserID == userID ||
					obj.InspectorUserID == userID {
					return true
				}
			}
		}
	}

	// 3. Проверяем через путь к файлу
	// Путь формата: {storageRoot}/{role}/{objectID}/documents/{filename}
	relPath, err := filepath.Rel(h.storageRoot, doc.StoragePath)
	if err == nil {
		parts := strings.Split(filepath.ToSlash(relPath), "/")
		if len(parts) >= 3 {
			// parts[0] = role (customer/foreman/inspector)
			// parts[1] = objectID
			objectID := parts[1]

			var obj models.Object
			if err := h.db.First(&obj, objectID).Error; err == nil {
				if obj.CustomerControlUserID == userID ||
					obj.ForemanUserID == userID ||
					obj.InspectorUserID == userID {
					return true
				}
			}
		}
	}

	// 4. Проверяем администратора
	var user models.User
	if err := h.db.First(&user, userID).Error; err == nil && user.Role == models.RoleAdmin {
		if user.Role == models.RoleAdmin {
			return true
		}
	}

	return false
}
