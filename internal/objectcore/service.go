package objectcore

import (
	"errors"
	"math"

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
		ID:                     obj.ID,
		Name:                   obj.Name,
		City:                   obj.City,
		Address:                obj.Address,
		Description:            obj.Description,
		Status:                 obj.Status,
		Lat:                    obj.Lat,
		Lng:                    obj.Lng,
		PlannedStartDate:       obj.PlannedStartDate,
		PlannedEndDate:         obj.PlannedEndDate,
		InitActFilePath:        obj.InitActFilePath,
		InitChecklistJSON:      obj.InitChecklistJSON,
		ActualStartDate:        obj.ActualStartDate,
		ActivationRejectReason: obj.ActivationRejectReason,
		Progress:               CalcProgress(db, obj.ID),
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

func BuildObjectDetailDTO(db *gorm.DB, obj *models.Object) ObjectDetailDTO {
	core := BuildObjectCoreDTO(db, obj)

	var items []models.WorkItem
	db.Where("object_id = ?", obj.ID).Order("id ASC").Find(&items)

	var deliveries []models.MaterialDelivery
	db.Where("object_id = ?", obj.ID).Order("date DESC, id DESC").Find(&deliveries)

	return ObjectDetailDTO{
		Object:     core,
		WorkItems:  items,
		Deliveries: deliveries,
	}
}

// calcProgress считает процент завершённых этапов (статус DONE).
// Если этапов нет совсем — возвращаем 0, не делим на ноль.
func CalcProgress(db *gorm.DB, objectID uint) float64 {
	type row struct {
		TotalPlan float64
		TotalFact float64
	}
	var r row
	db.Raw(`
        SELECT
            COALESCE(SUM(wi.plan_qty), 0) AS total_plan,
            COALESCE(SUM(wr_sum.fact), 0) AS total_fact
        FROM work_items wi
        LEFT JOIN (
            SELECT work_item_id, SUM(qty) AS fact
            FROM work_reports
            GROUP BY work_item_id
        ) wr_sum ON wr_sum.work_item_id = wi.id
        WHERE wi.object_id = ?
    `, objectID).Scan(&r)

	if r.TotalPlan == 0 {
		return 0
	}
	p := r.TotalFact / r.TotalPlan * 100
	if p > 100 {
		p = 100
	}
	return math.Round(p*10) / 10
}
