package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/safrizal-hk/uas-gofiber/app/service"
	"github.com/safrizal-hk/uas-gofiber/middleware"
)

func RegisterUserRoutes(v1 fiber.Router, userService *service.AdminManageUsersService) {
	
	const managePerm = "user:manage" 
	
	userRoute := v1.Group("/users", middleware.AuthRequired) 
	
	userRoute.Post("/", middleware.RBACRequired(managePerm), userService.AdminCreateUser) 
	userRoute.Get("/", middleware.RBACRequired(managePerm), userService.AdminListUsers) 
	userRoute.Get("/:id", middleware.RBACRequired(managePerm), userService.AdminGetUserByID)
	userRoute.Put("/:id", middleware.RBACRequired(managePerm), userService.AdminUpdateUser)
	userRoute.Delete("/:id", middleware.RBACRequired(managePerm), userService.AdminDeleteUser)
	userRoute.Put("/:id/role", middleware.RBACRequired(managePerm), userService.AdminSetUserRole)
	userRoute.Put("/students/:id/advisor", middleware.RBACRequired(managePerm), userService.AdminSetAdvisor) 
}