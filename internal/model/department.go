package model

import "time"

// Department представляет подразделение в организационной структуре
type Department struct {
    ID        uint       `json:"id" gorm:"primaryKey"`
    Name      string     `json:"name" gorm:"not null;size:200"`
    ParentID  *uint      `json:"parent_id" gorm:"index"`  // nil для корневого подразделения
    CreatedAt time.Time  `json:"created_at"`
    
    // Связи
    Employees []Employee   `json:"employees,omitempty" gorm:"foreignKey:DepartmentID"`
    Children  []Department `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}
