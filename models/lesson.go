package models

import "gorm.io/gorm"

type Lesson struct {
	gorm.Model
	Title    string `json:"title"`
	Content  string `json:"content"`
	CourseID uint   `json:"course_id"`
}
