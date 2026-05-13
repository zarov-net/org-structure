package model

import "time"

type Employee struct {
    ID           uint      `json:"id" gorm:"primaryKey"`
    DepartmentID uint      `json:"department_id" gorm:"not null;index"`
    FullName     string    `json:"full_name" gorm:"not null;size:200"`
    Position     string    `json:"position" gorm:"not null;size:200"`
    HiredAt      *string   `json:"hired_at"` // формат DATE, может быть null
    CreatedAt    time.Time `json:"created_at"`
    
    // Связь с подразделением (не отображается в JSON)
    Department Department `json:"-" gorm:"foreignKey:DepartmentID"`
}
