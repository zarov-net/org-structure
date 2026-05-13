package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"org-structure/internal/model"
	"org-structure/internal/repository"
	"org-structure/internal/service"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateDepartment(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.Department{}, &model.Employee{})

	repo := repository.NewDepartmentRepo(db)
	svc := service.NewDepartmentService(repo)
	h := NewDepartmentHandler(svc)

	body := map[string]interface{}{"name": "Тестовый отдел"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var dept model.Department
	json.Unmarshal(w.Body.Bytes(), &dept)
	if dept.Name != "Тестовый отдел" {
		t.Errorf("expected 'Тестовый отдел', got '%s'", dept.Name)
	}
	if dept.ID != 1 {
		t.Errorf("expected ID 1, got %d", dept.ID)
	}
}
