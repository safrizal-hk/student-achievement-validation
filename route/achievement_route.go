package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/safrizal-hk/uas-gofiber/app/service" 
	"github.com/safrizal-hk/uas-gofiber/middleware"
)

func RegisterAchievementRoutes(v1 fiber.Router, achievementService *service.AchievementService) {
	protected := v1.Group("/achievements", middleware.AuthRequired) 
	
	protected.Get("/", middleware.RBACRequired("achievement:read"), achievementService.ListAllAchievements)
	protected.Get("/:id", middleware.RBACRequired("achievement:read"), achievementService.GetAchievementDetail)
	protected.Post("/", middleware.RBACRequired("achievement:create"), achievementService.SubmitPrestasi)
	protected.Put("/:id", middleware.RBACRequired("achievement:update"), achievementService.UpdatePrestasi)
	protected.Delete("/:id", middleware.RBACRequired("achievement:delete"), achievementService.DeletePrestasi)
	protected.Post("/:id/submit", middleware.RBACRequired("achievement:update"), achievementService.SubmitForVerification)
	protected.Post("/:id/verify", middleware.RBACRequired("achievement:verify"), achievementService.VerifyPrestasi)
	protected.Post("/:id/reject", middleware.RBACRequired("achievement:verify"), achievementService.RejectPrestasi)
	protected.Get("/:id/history", middleware.RBACRequired("achievement:read"), achievementService.GetHistory)
	protected.Post("/:id/attachments", middleware.RBACRequired("achievement:update"), achievementService.AddAttachment)
}