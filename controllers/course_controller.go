package controllers

import (
	"github.com/gin-gonic/gin"
	"online-course-platform/models"
)

func CreateCourse(c *gin.Context) {
	var course models.Course
	if err := c.ShouldBindJSON(&course); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Create(&course).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, course)
}

func GetCourse(c *gin.Context) {
	id := c.Param("id")
	var course models.Course
	if err := DB.First(&course, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Course not found"})
		return
	}
	c.JSON(200, course)
}

func UpdateCourse(c *gin.Context) {
	id := c.Param("id")
	var course models.Course
	if err := DB.First(&course, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Course not found"})
		return
	}
	if err := c.ShouldBindJSON(&course); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	DB.Save(&course)
	c.JSON(200, course)
}

func DeleteCourse(c *gin.Context) {
	id := c.Param("id")
	if err := DB.Delete(&models.Course{}, id).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Course deleted"})
}
