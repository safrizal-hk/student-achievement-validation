package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/safrizal-hk/uas-gofiber/app/service" 
	"github.com/safrizal-hk/uas-gofiber/middleware"
)

// RegisterReportRoutes mendaftarkan rute untuk Reporting & Analytics (FR-011)
func RegisterReportRoutes(v1 fiber.Router, reportService *service.ReportService) {
	
	// Semua rute report memerlukan otentikasi
	reportsRoute := v1.Group("/reports", middleware.AuthRequired)

	// GET /api/v1/reports/statistics (FR-011)
	reportsRoute.Get("/statistics", middleware.RBACRequired("achievement:read"), reportService.GetStatisticsHandler)
	
	// GET /api/v1/reports/student/:id (Report Detail Mahasiswa)
	reportsRoute.Get("/student/:id", middleware.RBACRequired("achievement:read"), reportService.GetStudentReportHandler)
}