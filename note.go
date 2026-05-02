package models

import (
	"time"

	"gorm.io/gorm"
)

type Note struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"not null"`
	Description string         `json:"description"`
	FilePath    string         `json:"-" gorm:"not null"`
	FileName    string         `json:"file_name" gorm:"not null"`
	FileSize    int64          `json:"file_size"`
	FileType    string         `json:"file_type"`
	CourseID    uint           `json:"course_id" gorm:"not null"`
	UserID      uint           `json:"user_id" gorm:"not null"`
	Downloads   int            `json:"downloads" gorm:"default:0"`
	IsPublic    bool           `json:"is_public" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User      User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Course    Course     `json:"course,omitempty" gorm:"foreignKey:CourseID"`
	TodoLists []TodoList `json:"todo_lists,omitempty" gorm:"foreignKey:NoteID"`
}

type CreateNoteRequest struct {
	Title       string `form:"title" binding:"required,min=2,max=200"`
	Description string `form:"description"`
	CourseID    uint   `form:"course_id" binding:"required"`
	IsPublic    bool   `form:"is_public"`
}

type UpdateNoteRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	IsPublic    *bool  `json:"is_public"`
}