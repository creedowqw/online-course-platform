package controllers

import (
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"log"
	"net/http"
	"online-course-platform/internal/models"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	DB *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterEmailInput struct {
	Email string `json:"email"`
}

type RoleUpdateInput struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

func generateJWT(id uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"id":   id,
		"role": role,
		"exp":  time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func (uc *UserController) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Неверный формат данных"})
		return
	}

	var user models.User
	if err := uc.DB.Where("name = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "Пользователь не найден"})
		return
	}

	token, err := generateJWT(user.ID, user.Role)
	if err != nil {
		c.JSON(500, gin.H{"error": "Ошибка генерации токена"})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
	})
}

func (uc *UserController) RegisterOrLogin(c *gin.Context) {
	var input RegisterEmailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный запрос"})
		return
	}

	email := strings.ToLower(input.Email)
	if !strings.HasSuffix(email, "@narxoz.kz") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Только корпоративные email @narxoz.kz"})
		return
	}

	var user models.User
	err := uc.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		user = models.User{
			Email:    email,
			Name:     strings.Split(email, "@")[0],
			Role:     "student",
			Password: "",
		}
		uc.DB.Create(&user)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}

func (uc *UserController) UpdateUserRole(c *gin.Context) {
	var input RoleUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	var user models.User
	if err := uc.DB.Where("name = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	user.Role = input.Role
	uc.DB.Save(&user)
	c.JSON(http.StatusOK, gin.H{"message": "Роль обновлена"})
}

func EnsureAdminExists(db *gorm.DB) models.User {
	var admin models.User
	if err := db.Where("email = ?", "admin@narxoz.kz").First(&admin).Error; err != nil {
		admin = models.User{Name: "admin", Email: "admin@narxoz.kz", Role: "admin"}
		db.Create(&admin)
		log.Println("Админ создан: admin@narxoz.kz / admin")
	}
	return admin
}
