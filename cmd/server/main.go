package main

import (
	"log"
	"net/http"
	"strings"

	"org-structure/internal/config"
	"org-structure/internal/handler"
	"org-structure/internal/middleware"
	"org-structure/internal/model"
	"org-structure/internal/repository"
	"org-structure/internal/service"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "org-structure/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title API организационной структуры
// @version 1.0
// @description REST API для управления организационной структурой компании
// @host localhost:8080
// @BasePath /

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}

	err = db.AutoMigrate(&model.Department{}, &model.Employee{})
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
	log.Println("Migrations completed")

	deptRepo := repository.NewDepartmentRepo(db)
	empRepo := repository.NewEmployeeRepo(db)

	deptService := service.NewDepartmentService(deptRepo)
	empService := service.NewEmployeeService(empRepo, deptRepo)

	deptHandler := handler.NewDepartmentHandler(deptService)
	empHandler := handler.NewEmployeeHandler(empService)

	mux := http.NewServeMux()

	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/departments", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/departments" && r.Method == http.MethodPost {
			deptHandler.Create(w, r)
			return
		}

		if strings.HasPrefix(path, "/departments/") {
			handleDepartmentRoute(w, r, deptHandler, empHandler)
			return
		}

		http.NotFound(w, r)
	})

	mux.HandleFunc("/departments/", func(w http.ResponseWriter, r *http.Request) {
		handleDepartmentRoute(w, r, deptHandler, empHandler)
	})

	h := middleware.Logging(mux)
	h = middleware.Recovery(h)

	log.Printf("Server started on :%s", cfg.ServerPort)
	log.Printf("Swagger: http://localhost:%s/swagger/index.html", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, h))
}

func handleDepartmentRoute(w http.ResponseWriter, r *http.Request, deptHandler *handler.DepartmentHandler, empHandler *handler.EmployeeHandler) {
	path := strings.TrimPrefix(r.URL.Path, "/departments/")
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) >= 2 && parts[1] == "employees" {
		if r.Method == http.MethodPost {
			empHandler.Create(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch r.Method {
	case http.MethodGet:
		deptHandler.GetByID(w, r)
	case http.MethodPatch:
		deptHandler.Update(w, r)
	case http.MethodDelete:
		deptHandler.Delete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
