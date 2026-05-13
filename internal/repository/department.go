package repository

import (
	"errors"
	"fmt"

	"org-structure/internal/model"

	"gorm.io/gorm"
)

type DepartmentRepo struct {
	db *gorm.DB
}

func NewDepartmentRepo(db *gorm.DB) *DepartmentRepo {
	return &DepartmentRepo{db: db}
}

// Create создает новое подразделение с проверкой уникальности имени
func (r *DepartmentRepo) Create(department *model.Department) error {
	// Проверяем уникальность имени в рамках одного родителя
	var count int64
	query := r.db.Model(&model.Department{}).Where("name = ?", department.Name)

	if department.ParentID != nil {
		query = query.Where("parent_id = ?", *department.ParentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	if err := query.Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return errors.New("подразделение с таким названием уже существует в этом родителе")
	}

	return r.db.Create(department).Error
}

// GetByID получает подразделение по ID с сотрудниками
func (r *DepartmentRepo) GetByID(id uint) (*model.Department, error) {
	var department model.Department
	err := r.db.Preload("Employees").First(&department, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("подразделение с ID %d не найдено", id)
		}
		return nil, err
	}
	return &department, nil
}

// GetEmployees получает сотрудников подразделения с сортировкой
func (r *DepartmentRepo) GetEmployees(departmentID uint) ([]model.Employee, error) {
	var employees []model.Employee
	err := r.db.Where("department_id = ?", departmentID).
		Order("created_at ASC, full_name ASC").
		Find(&employees).Error
	if err != nil {
		return nil, err
	}
	if employees == nil {
		employees = []model.Employee{}
	}
	return employees, err
}

// GetChildren рекурсивно получает дочерние подразделения
func (r *DepartmentRepo) GetChildren(parentID uint, depth int, includeEmployees bool) ([]model.Department, error) {
	if depth <= 0 {
		return []model.Department{}, nil
	}

	var departments []model.Department
	query := r.db.Where("parent_id = ?", parentID)

	if includeEmployees {
		query = query.Preload("Employees")
	}

	if err := query.Find(&departments).Error; err != nil {
		return nil, err
	}

	// Рекурсивно получаем детей для каждого найденного подразделения
	for i := range departments {
		children, err := r.GetChildren(departments[i].ID, depth-1, includeEmployees)
		if err != nil {
			return nil, err
		}
		if children == nil {
			children = []model.Department{}
		}
		departments[i].Children = children
	}

	return departments, nil
}

// Update обновляет только указанные поля подразделения
func (r *DepartmentRepo) Update(id uint, updates map[string]interface{}) error {
	// Проверяем существование
	var existing model.Department
	if err := r.db.First(&existing, id).Error; err != nil {
		return err
	}

	// Проверяем уникальность имени если оно меняется
	if name, ok := updates["name"]; ok {
		var count int64
		query := r.db.Model(&model.Department{}).Where("name = ? AND id != ?", name, id)

		if parentID, hasParent := updates["parent_id"]; hasParent && parentID != nil {
			query = query.Where("parent_id = ?", parentID)
		} else if hasParent && parentID == nil {
			query = query.Where("parent_id IS NULL")
		} else if existing.ParentID != nil {
			query = query.Where("parent_id = ?", *existing.ParentID)
		} else {
			query = query.Where("parent_id IS NULL")
		}

		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("подразделение с таким названием уже существует в этом родителе")
		}
	}

	return r.db.Model(&model.Department{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteCascade удаляет подразделение со всеми потомками и сотрудниками
func (r *DepartmentRepo) DeleteCascade(id uint) error {
	// Загружаем все дочерние ID
	ids, err := r.GetDescendantIDs(id)
	if err != nil {
		return err
	}

	// Удаляем в обратном порядке (сначала самые глубокие)
	for i := len(ids) - 1; i >= 0; i-- {
		if err := r.db.Unscoped().Delete(&model.Department{}, ids[i]).Error; err != nil {
			return err
		}
	}

	return nil
}

// DeleteReassign удаляет подразделение и переназначает сотрудников
func (r *DepartmentRepo) DeleteReassign(id uint, reassignToID uint) error {
	tx := r.db.Begin()

	// Переносим сотрудников в новое подразделение
	if err := tx.Model(&model.Employee{}).Where("department_id = ?", id).Update("department_id", reassignToID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Удаляем подразделение (дочерние останутся с parent_id = null)
	if err := tx.Delete(&model.Department{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetDescendantIDs получает все ID потомков подразделения
func (r *DepartmentRepo) GetDescendantIDs(parentID uint) ([]uint, error) {
	var ids []uint

	// Добавляем сам ID родителя
	ids = append(ids, parentID)

	// Рекурсивно получаем всех потомков
	children, err := r.GetChildrenIDs(parentID)
	if err != nil {
		return nil, err
	}
	ids = append(ids, children...)

	return ids, nil
}

// GetChildrenIDs получает ID всех непосредственных детей
func (r *DepartmentRepo) GetChildrenIDs(parentID uint) ([]uint, error) {
	var ids []uint

	var children []model.Department
	if err := r.db.Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return nil, err
	}

	for _, child := range children {
		ids = append(ids, child.ID)
		// Рекурсивно получаем детей детей
		childIDs, err := r.GetChildrenIDs(child.ID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, childIDs...)
	}

	return ids, nil
}
