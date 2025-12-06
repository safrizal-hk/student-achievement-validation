package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/safrizal-hk/uas-gofiber/app/service" 
	"github.com/safrizal-hk/uas-gofiber/middleware"
)

func RegisterReportRoutes(v1 fiber.Router, reportService *service.ReportService) {
	reportsRoute := v1.Group("/reports", middleware.AuthRequired)

	reportsRoute.Get("/statistics", middleware.RBACRequired("achievement:read"), reportService.GetStatistics)
	reportsRoute.Get("/student/:id", middleware.RBACRequired("achievement:read"), reportService.GetStudentReport)
}