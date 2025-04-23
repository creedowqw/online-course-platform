package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"online-course-platform/internal/models"
)

type CourseController struct {
	DB *gorm.DB
}

func NewCourseController(db *gorm.DB) *CourseController {
	return &CourseController{DB: db}
}

func (cc *CourseController) CreateCourse(c *gin.Context) {
	var course models.Course
	if err := c.ShouldBindJSON(&course); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := cc.DB.Create(&course).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, course)
}

func (cc *CourseController) GetCourse(c *gin.Context) {
	id := c.Param("id")
	var course models.Course
	if err := cc.DB.First(&course, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Course not found"})
		return
	}
	c.JSON(200, course)
}

func (cc *CourseController) UpdateCourse(c *gin.Context) {
	id := c.Param("id")
	var course models.Course
	if err := cc.DB.First(&course, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Course not found"})
		return
	}
	if err := c.ShouldBindJSON(&course); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	cc.DB.Save(&course)
	c.JSON(200, course)
}

func (cc *CourseController) DeleteCourse(c *gin.Context) {
	id := c.Param("id")
	if err := cc.DB.Delete(&models.Course{}, id).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Course deleted"})
}

func (cc *CourseController) GetLessons(context *gin.Context) {

}

func (cc *CourseController) GetAllCourses(c *gin.Context) {
	var courses []models.Course
	if err := cc.DB.Find(&courses).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, courses)
}
