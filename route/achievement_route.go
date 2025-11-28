package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/safrizal-hk/uas-gofiber/app/service" 
	"github.com/safrizal-hk/uas-gofiber/middleware"      
)

func RegisterAchievementRoutes(v1 fiber.Router, achievementService *service.AchievementService) {
	protected := v1.Group("/achievements", middleware.AuthRequired) 
	
	// FR-003: Submit Prestasi (Draft)
	protected.Post("/", middleware.RBACRequired("achievement:create"), achievementService.SubmitPrestasi)
	
	// FR-004: Submit for Verification
	protected.Post("/:id/submit", middleware.RBACRequired("achievement:update"), achievementService.SubmitForVerification)
	
	// FR-005: Soft Delete Prestasi (Draft Only)
	protected.Delete("/:id", middleware.RBACRequired("achievement:delete"), achievementService.DeletePrestasi)
	
	// FR-006 & FR-010: List Prestasi (Filtered by Role)
	protected.Get("/", middleware.RBACRequired("achievement:read"), achievementService.ListAllAchievements)

	// FR-007: Verify Prestasi (Lecturer/Admin)
	protected.Post("/:id/verify", middleware.RBACRequired("achievement:verify"), achievementService.VerifyPrestasi)
	
	// FR-008: Reject Prestasi (Lecturer/Admin)
	protected.Post("/:id/reject", middleware.RBACRequired("achievement:verify"), achievementService.RejectPrestasi)
}