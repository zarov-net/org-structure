package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"org-structure/internal/model"
	"org-structure/internal/service"
)

type DepartmentHandler struct {
	service *service.DepartmentService
}

func NewDepartmentHandler(service *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: service}
}

// Create создает новое подразделение
// @Summary Создать подразделение
// @Description Создает новое подразделение в организационной структуре
// @Tags Подразделения
// @Accept json
// @Produce json
// @Param department body model.Department true "Данные подразделения"
// @Success 201 {object} model.Department "Подразделение создано"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Router /departments [post]
func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var department model.Department
	if err := json.NewDecoder(r.Body).Decode(&department); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректные данные"})
		return
	}
	if err := h.service.ValidateAndCreate(&department); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, department)
}

// GetByID получает подразделение по ID
// @Summary Получить подразделение
// @Description Возвращает подразделение с сотрудниками и дочерними подразделениями
// @Tags Подразделения
// @Produce json
// @Param id path int true "ID подразделения"
// @Param depth query int false "Глубина дерева (1-5)" default(1) minimum(1) maximum(5)
// @Param include_employees query bool false "Включать сотрудников" default(true)
// @Success 200 {object} map[string]interface{} "Данные подразделения"
// @Failure 404 {object} map[string]string "Подразделение не найдено"
// @Router /departments/{id} [get]
func (h *DepartmentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный ID"})
		return
	}

	depthStr := r.URL.Query().Get("depth")
	depth := 1
	if depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil && d >= 1 && d <= 5 {
			depth = d
		}
	}

	includeEmployees := r.URL.Query().Get("include_employees") != "false"

	result, err := h.service.GetByIDWithTree(id, depth, includeEmployees)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "подразделение не найдено"})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// Update обновляет подразделение
// @Summary Обновить подразделение
// @Description Изменяет название или родителя подразделения
// @Tags Подразделения
// @Accept json
// @Produce json
// @Param id path int true "ID подразделения"
// @Param department body model.Department true "Новые данные"
// @Success 200 {object} model.Department "Подразделение обновлено"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 409 {object} map[string]string "Конфликт (цикл в дереве)"
// @Router /departments/{id} [patch]
func (h *DepartmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный ID"})
		return
	}

	var department model.Department
	if err := json.NewDecoder(r.Body).Decode(&department); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректные данные"})
		return
	}
	department.ID = id

	if err := h.service.ValidateAndUpdate(&department); err != nil {
		if strings.Contains(err.Error(), "цикл") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		} else {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return
	}
	writeJSON(w, http.StatusOK, department)
}

// Delete удаляет подразделение
// @Summary Удалить подразделение
// @Description Удаляет подразделение. Режимы: cascade (каскадное удаление) или reassign (перенос сотрудников)
// @Tags Подразделения
// @Produce json
// @Param id path int true "ID подразделения"
// @Param mode query string true "Режим удаления" Enums(cascade, reassign)
// @Param reassign_to_department_id query int false "ID подразделения для переноса сотрудников (обязателен при mode=reassign)"
// @Success 204 "Подразделение удалено"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Router /departments/{id} [delete]
func (h *DepartmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некорректный ID"})
		return
	}

	mode := r.URL.Query().Get("mode")
	reassignTo := r.URL.Query().Get("reassign_to_department_id")

	if err := h.service.Delete(id, mode, reassignTo); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func extractID(path string) (uint, error) {
	path = strings.TrimPrefix(path, "/departments/")
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		return 0, strconv.ErrSyntax
	}
	id, err := strconv.ParseUint(parts[0], 10, 32)
	return uint(id), err
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
