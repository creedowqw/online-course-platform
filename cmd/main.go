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
	dsn := "host=db user=user password=password dbname=online_courses port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	db.AutoMigrate(&models.User{}, &models.Course{}, &models.Lesson{})
	controllers.InitDatabase(db)

	r := gin.Default()
	r.POST("/courses", controllers.CreateCourse)
	r.GET("/courses/:id", controllers.GetCourse)
	r.PUT("/courses/:id", controllers.UpdateCourse)
	r.DELETE("/courses/:id", controllers.DeleteCourse)

	r.POST("/lessons", controllers.CreateLesson)
	r.GET("/courses/:course_id/lessons", controllers.GetLessons)

	r.Run(":8080")
}
