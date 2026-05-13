package service

import (
	"errors"
	"org-structure/internal/model"
	"org-structure/internal/repository"
	"strings"
)

type EmployeeService struct {
	empRepo  *repository.EmployeeRepo
	deptRepo *repository.DepartmentRepo
}

func NewEmployeeService(empRepo *repository.EmployeeRepo, deptRepo *repository.DepartmentRepo) *EmployeeService {
	return &EmployeeService{empRepo: empRepo, deptRepo: deptRepo}
}

func (s *EmployeeService) Create(employee *model.Employee) error {
	employee.FullName = strings.TrimSpace(employee.FullName)
	employee.Position = strings.TrimSpace(employee.Position)

	if employee.FullName == "" {
		return errors.New("полное имя не может быть пустым")
	}
	if len(employee.FullName) > 200 {
		return errors.New("полное имя не может быть длиннее 200 символов")
	}
	if employee.Position == "" {
		return errors.New("должность не может быть пустой")
	}
	if len(employee.Position) > 200 {
		return errors.New("должность не может быть длиннее 200 символов")
	}

	_, err := s.deptRepo.GetByID(employee.DepartmentID)
	if err != nil {
		return errors.New("подразделение не найдено")
	}

	return s.empRepo.Create(employee)
}
