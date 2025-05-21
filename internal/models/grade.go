package models

import "gorm.io/gorm"

type Grade struct {
	gorm.Model
	UserID   uint `json:"user_id"`
	CourseID uint `json:"course_id"`
	Score    int  `json:"score"`
}
