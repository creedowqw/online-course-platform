package models

import "gorm.io/gorm"

type Course struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	TeacherID   uint   `json:"teacher_id"`
}
