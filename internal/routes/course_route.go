package routes

import (
	"github.com/gin-gonic/gin"
	"online-course-platform/internal/controllers"
)

func SetupCourseRoutes(controller controllers.CourseController) *gin.Engine {
	r := gin.Default()

	courses := r.Group("/courses")
	{
		courses.GET("", controller.GetAllCourses)
		courses.POST("", controller.CreateCourse)
		courses.GET("/:id", controller.GetCourse)
		courses.PUT("/:id", controller.UpdateCourse)
		courses.DELETE("/:id", controller.DeleteCourse)
		courses.GET("/:id/lessons", controller.GetLessons)
	}

	return r
}
