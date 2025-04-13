package controllers

import (
	"gorm.io/gorm"
	"log"
	"online-course-platform/models"
)

var DB *gorm.DB

func InitDatabase(database *gorm.DB) {
	DB = database
}

func CreateDefaultAdmin() {
	var count int64
	if err := DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count).Error; err != nil {
		log.Println("Ошибка при проверке админа:", err)
		return
	}

	if count == 0 {
		admin := models.User{
			Name:     "admin",
			Password: "admin123",
			Role:     "admin",
		}
		if err := DB.Create(&admin).Error; err != nil {
			log.Println("Ошибка при создании админа:", err)
		} else {
			log.Println("Админ успешно создан: admin / admin123")
		}
	}
}
