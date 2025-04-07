package controllers

import "gorm.io/gorm"

var DB *gorm.DB

func InitDatabase(database *gorm.DB) {
	DB = database
}
