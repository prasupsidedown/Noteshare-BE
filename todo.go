package models

import (
	"time"

	"gorm.io/gorm"
)

type TodoItem struct {
	Task      string `json:"task"`
	Priority  string `json:"priority"` // high, medium, low
	Completed bool   `json:"completed"`
}

type TodoList struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	NoteID    uint           `json:"note_id" gorm:"not null"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Title     string         `json:"title"`
	Items     []TodoItem     `json:"items" gorm:"serializer:json"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Note Note `json:"note,omitempty" gorm:"foreignKey:NoteID"`
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type UpdateTodoItemRequest struct {
	Index     int  `json:"index" binding:"required"`
	Completed bool `json:"completed"`
}