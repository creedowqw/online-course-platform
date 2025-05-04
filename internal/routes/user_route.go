package routes

import (
	"github.com/gin-gonic/gin"
	"online-course-platform/internal/controllers"
)

func RegisterUserRoutes(r *gin.RouterGroup, controller controllers.UserController) {
	r.POST("/login", controller.Login)
	r.POST("/register-or-login", controller.RegisterOrLogin)
	r.POST("/update-role", controller.UpdateUserRole)
}
