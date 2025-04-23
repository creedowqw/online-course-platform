package routes

import (
	"github.com/gin-gonic/gin"
	"online-course-platform/internal/controllers"
)

func SetupUserRoutes(controller controllers.UserController) *gin.Engine {
	r := gin.Default()

	userRoutes := r.Group("/courses")
	{
		userRoutes.POST("/login", controller.Login)
		userRoutes.POST("/register-or-login", controller.RegisterOrLogin)
		userRoutes.POST("/update-role", controller.UpdateUserRole)
	}

	return r
}
