package objectcore

import (
	"errors"

	"github.com/Loc-Leonard/Diploma/internal/models"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("object not found")
var ErrForbidden = errors.New("access denied")

// Загружает объект и проверяет доступ по роли пользователя
func LoadObjectForUser(db *gorm.DB, objectID string, userID uint, role string) (*models.Object, error) {
	var obj models.Object
	if err := db.First(&obj, objectID).Error; err != nil {
		return nil, ErrNotFound
	}

	switch role {
	case string(models.RoleCustomer):
		if obj.CustomerControlUserID != userID {
			return nil, ErrForbidden
		}
	case string(models.RoleForeman):
		if obj.ForemanUserID != userID {
			return nil, ErrForbidden
		}
	case string(models.RoleInspector):
		if obj.InspectorUserID != userID {
			return nil, ErrForbidden
		}
	default:
		return nil, ErrForbidden
	}

	return &obj, nil
}

// Строит базовое DTO объекта, одинаковое для всех ролей
func BuildObjectCoreDTO(db *gorm.DB, obj *models.Object) ObjectCoreDTO {
	dto := ObjectCoreDTO{
		ID:               obj.ID,
		Name:             obj.Name,
		City:             obj.City,
		Address:          obj.Address,
		Description:      obj.Description,
		Status:           obj.Status,
		Lat:              obj.Lat,
		Lng:              obj.Lng,
		PlannedStartDate: obj.PlannedStartDate,
		PlannedEndDate:   obj.PlannedEndDate,
	}

	loadPerson := func(id uint) *ObjectPersonDTO {
		if id == 0 {
			return nil
		}
		var u models.User
		if err := db.First(&u, id).Error; err != nil {
			return nil
		}
		return &ObjectPersonDTO{ID: u.ID, FullName: u.FullName}
	}

	dto.Customer = loadPerson(obj.CustomerControlUserID)
	dto.Foreman = loadPerson(obj.ForemanUserID)
	dto.Inspector = loadPerson(obj.InspectorUserID)

	return dto
}
