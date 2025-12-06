package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/safrizal-hk/uas-gofiber/app/service"
	"github.com/safrizal-hk/uas-gofiber/middleware"
)

func RegisterStudentRoutes(v1 fiber.Router, studentService *service.StudentService) {
	studentRoute := v1.Group("/students", middleware.AuthRequired)

	studentRoute.Get("/", middleware.RBACRequired("user:manage"), studentService.ListStudents)
	studentRoute.Get("/:id", middleware.RBACRequired("user:manage"), studentService.GetStudentDetail)
	studentRoute.Put("/:id/advisor", middleware.RBACRequired("user:manage"), studentService.SetAdvisor)
	studentRoute.Get("/:id/achievements", middleware.RBACRequired("user:manage"), studentService.GetStudentAchievements)
}