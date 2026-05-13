package repository

import (
	"errors"
	"fmt"
	"org-structure/internal/model"

	"gorm.io/gorm"
)

type EmployeeRepo struct {
	db *gorm.DB
}

func NewEmployeeRepo(db *gorm.DB) *EmployeeRepo {
	return &EmployeeRepo{db: db}
}

func (r *EmployeeRepo) Create(employee *model.Employee) error {
	var dept model.Department
	if err := r.db.First(&dept, employee.DepartmentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("подразделение с ID %d не найдено", employee.DepartmentID)
		}
		return err
	}
	return r.db.Create(employee).Error
}

func (r *EmployeeRepo) GetByDepartmentID(departmentID uint) ([]model.Employee, error) {
	var employees []model.Employee
	err := r.db.Where("department_id = ?", departmentID).
		Order("created_at ASC, full_name ASC").
		Find(&employees).Error
	return employees, err
}
