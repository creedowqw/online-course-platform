package controllers

import (
	"github.com/gin-gonic/gin"
	"online-course-platform/models"
)

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Неверный формат данных"})
		return
	}

	var user models.User
	if err := DB.Where("name = ? AND password = ?", input.Username, input.Password).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "Неверный логин или пароль"})
		return
	}
	c.JSON(200, gin.H{
		"id":   user.ID,
		"name": user.Name,
		"role": user.Role,
	})
}

func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if user.Name == "" || user.Password == "" || user.Role == "" {
		c.JSON(400, gin.H{"error": "name, password и role обязательны"})
		return
	}

	if err := DB.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, user)
}
