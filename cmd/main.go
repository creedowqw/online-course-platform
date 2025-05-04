package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"online-course-platform/internal/bot"
	"online-course-platform/internal/controllers"
	"online-course-platform/internal/db"
	"online-course-platform/internal/routes"
)

func main() {
	godotenv.Load()
	db.InitDB()

	controllers.EnsureAdminExists(db.DB)
	go bot.StartBot()

	courseController := *controllers.NewCourseController(db.DB)
	userController := *controllers.NewUserController(db.DB)

	r := gin.Default()

	courseGroup := r.Group("/courses")
	routes.RegisterCourseRoutes(courseGroup, courseController)

	userGroup := r.Group("/users")
	routes.RegisterUserRoutes(userGroup, userController)

	r.Run(":8080")
}
