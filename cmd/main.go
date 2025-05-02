package main

import (
	"github.com/joho/godotenv"
	"log"
	"online-course-platform/internal/bot"
	"online-course-platform/internal/controllers"
	"online-course-platform/internal/db"
	"online-course-platform/internal/models"
	"online-course-platform/internal/routes"
)

func main() {
	godotenv.Load()
	db.InitDB()

	var existingAdmin models.User
	if err := db.DB.Where("email = ?", "admin@narxoz.kz").First(&existingAdmin).Error; err != nil {
		db.DB.Create(&models.User{
			Name:  "admin",
			Email: "admin@narxoz.kz",
			Role:  "admin",
		})
		log.Println("Админ создан: admin@narxoz.kz / admin")
	}

	go bot.StartBot()

	courseController := *controllers.NewCourseController(db.DB)
	r := routes.SetupCourseRoutes(courseController)

	r.Run(":8080")

}
