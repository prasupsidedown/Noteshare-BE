package models

import (
	"time"

	"gorm.io/gorm"
)

type Course struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Code        string         `json:"code" gorm:"not null"`
	Description string         `json:"description"`
	Semester    string         `json:"semester"`
	UserID      uint           `json:"user_id" gorm:"not null"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User  User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Notes []Note `json:"notes,omitempty" gorm:"foreignKey:CourseID"`
}

type CreateCourseRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Code        string `json:"code" binding:"required,min=2,max=20"`
	Description string `json:"description"`
	Semester    string `json:"semester"`
}

type UpdateCourseRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Semester    string `json:"semester"`
}