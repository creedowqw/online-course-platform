package routes

import (
	"github.com/gin-gonic/gin"
	"online-course-platform/internal/controllers"
)

func RegisterCourseRoutes(r *gin.RouterGroup, controller controllers.CourseController) {
	r.GET("", controller.GetAllCourses)
	r.POST("", controller.CreateCourse)
	r.GET("/:id", controller.GetCourse)
	r.PUT("/:id", controller.UpdateCourse)
	r.DELETE("/:id", controller.DeleteCourse)
	r.GET("/:id/lessons", controller.GetLessons)
}
