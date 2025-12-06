package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/safrizal-hk/uas-gofiber/app/service"
	"github.com/safrizal-hk/uas-gofiber/middleware"
)

func RegisterLecturerRoutes(v1 fiber.Router, lecturerService *service.LecturerService) {
	lecturerRoute := v1.Group("/lecturers", middleware.AuthRequired)
	
	lecturerRoute.Get("/", lecturerService.ListLecturers)
	lecturerRoute.Get("/:id/advisees", middleware.RBACRequired("achievement:read"), lecturerService.GetLecturerAdvisees)
}