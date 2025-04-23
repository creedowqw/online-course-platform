package main

import (
	"github.com/joho/godotenv"
	"online-course-platform/internal/bot"
	"online-course-platform/internal/controllers"
	"online-course-platform/internal/db"
	"online-course-platform/internal/routes"
)

func main() {
	godotenv.Load()
	db.InitDB()

	go bot.StartBot()

	courseController := *controllers.NewCourseController(db.DB)
	r := routes.SetupCourseRoutes(courseController)

	r.Run(":8080")

}
