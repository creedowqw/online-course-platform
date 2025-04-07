package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"online-course-platform/models"
)

func CreateLesson(c *gin.Context) {
	var lesson models.Lesson
	if err := c.ShouldBindJSON(&lesson); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := db.Create(&lesson).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, lesson)
}

func GetLessons(c *gin.Context) {
	courseID := c.Param("course_id")
	var lessons []models.Lesson
	if err := db.Where("course_id = ?", courseID).Find(&lessons).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, lessons)
}
