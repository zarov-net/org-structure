package service

import (
	"errors"
	"fmt"
	"org-structure/internal/model"
	"org-structure/internal/repository"
	"strconv"
	"strings"
)

type DepartmentService struct {
	repo *repository.DepartmentRepo
}

func NewDepartmentService(repo *repository.DepartmentRepo) *DepartmentService {
	return &DepartmentService{repo: repo}
}

func (s *DepartmentService) ValidateAndCreate(department *model.Department) error {
	department.Name = strings.TrimSpace(department.Name)
	if department.Name == "" {
		return errors.New("название подразделения не может быть пустым")
	}
	if len(department.Name) > 200 {
		return errors.New("название подразделения не может быть длиннее 200 символов")
	}
	if department.ParentID != nil {
		_, err := s.repo.GetByID(*department.ParentID)
		if err != nil {
			return errors.New("родительское подразделение не найдено")
		}
	}
	return s.repo.Create(department)
}

func (s *DepartmentService) GetByIDWithTree(id uint, depth int, includeEmployees bool) (map[string]interface{}, error) {
	department, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{"department": department}
	if includeEmployees {
		employees, _ := s.repo.GetEmployees(id)
		if employees == nil {
			employees = []model.Employee{}
		}
		result["employees"] = employees
	}
	if depth > 0 {
		children, _ := s.repo.GetChildren(id, depth, includeEmployees)
		if children == nil {
			children = []model.Department{}
		}
		result["children"] = children
	}
	return result, nil
}

func (s *DepartmentService) ValidateAndUpdate(department *model.Department) error {
	if department.Name != "" {
		department.Name = strings.TrimSpace(department.Name)
	}
	existing, err := s.repo.GetByID(department.ID)
	if err != nil {
		return errors.New("подразделение не найдено")
	}
	if department.Name == "" {
		department.Name = existing.Name
	}
	if len(department.Name) > 200 {
		return errors.New("название подразделения не может быть длиннее 200 символов")
	}
	if department.ParentID == nil && existing.ParentID != nil {
		department.ParentID = existing.ParentID
	}
	if department.ParentID != nil {
		if *department.ParentID == department.ID {
			return errors.New("нельзя сделать подразделение родителем самого себя")
		}
		_, err := s.repo.GetByID(*department.ParentID)
		if err != nil {
			return errors.New("родительское подразделение не найдено")
		}
		if s.isDescendant(department.ID, *department.ParentID) {
			return errors.New("нельзя создать цикл в дереве подразделений")
		}
	}
	updates := map[string]interface{}{"name": department.Name}
	if department.ParentID != nil {
		updates["parent_id"] = *department.ParentID
	} else {
		updates["parent_id"] = nil
	}
	return s.repo.Update(department.ID, updates)
}

func (s *DepartmentService) Delete(id uint, mode string, reassignToID string) error {
	switch mode {
	case "cascade":
		return s.repo.DeleteCascade(id)
	case "reassign":
		if reassignToID == "" {
			return errors.New("не указан reassign_to_department_id")
		}
		reassignID, err := strconv.ParseUint(reassignToID, 10, 32)
		if err != nil {
			return errors.New("некорректный reassign_to_department_id")
		}
		_, err = s.repo.GetByID(uint(reassignID))
		if err != nil {
			return errors.New("целевое подразделение не найдено")
		}
		if s.isDescendant(id, uint(reassignID)) {
			return errors.New("нельзя переместить сотрудников в подчиненное подразделение")
		}
		return s.repo.DeleteReassign(id, uint(reassignID))
	default:
		return fmt.Errorf("неподдерживаемый режим удаления: %s", mode)
	}
}

func (s *DepartmentService) isDescendant(parentID, childID uint) bool {
	descendants, err := s.repo.GetDescendantIDs(parentID)
	if err != nil {
		return false
	}
	for _, id := range descendants {
		if id == childID {
			return true
		}
	}
	return false
}
