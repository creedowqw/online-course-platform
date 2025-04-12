package main

import (
	"log"
	"online-course-platform/controllers"
	"online-course-platform/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=user password=password dbname=online_courses port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	db.AutoMigrate(&models.User{}, &models.Course{}, &models.Lesson{})
	controllers.InitDatabase(db)

	r := gin.Default()
	courses := r.Group("/courses")
	{
		courses.POST("", controllers.CreateCourse)
		courses.GET("/:id", controllers.GetCourse)
		courses.PUT("/:id", controllers.UpdateCourse)
		courses.DELETE("/:id", controllers.DeleteCourse)
		courses.GET("/:id/lessons", controllers.GetLessons)

	}

	r.POST("/lessons", controllers.CreateLesson)
	r.Run(":8080")
}
