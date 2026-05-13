package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"org-structure/internal/model"
	"org-structure/internal/service"
)

type EmployeeHandler struct {
	service *service.EmployeeService
}

func NewEmployeeHandler(service *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: service}
}

// Create создает нового сотрудника
// @Summary Создать сотрудника
// @Description Создает нового сотрудника в указанном подразделении
// @Tags Сотрудники
// @Accept json
// @Produce json
// @Param id path int true "ID подразделения"
// @Param employee body model.Employee true "Данные сотрудника"
// @Success 201 {object} model.Employee "Сотрудник создан"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 404 {object} map[string]string "Подразделение не найдено"
// @Router /departments/{id}/employees [post]
func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/departments/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 || parts[1] != "employees" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "не найден"})
		return
	}

	deptID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный ID подразделения"})
		return
	}

	var employee model.Employee
	if err := json.NewDecoder(r.Body).Decode(&employee); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректные данные"})
		return
	}

	employee.DepartmentID = uint(deptID)

	if err := h.service.Create(&employee); err != nil {
		if strings.Contains(err.Error(), "не найдено") {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		} else {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return
	}

	writeJSON(w, http.StatusCreated, employee)
}
