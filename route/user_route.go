package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/safrizal-hk/uas-gofiber/app/service" // Mengakses AdminManageUsersService
	"github.com/safrizal-hk/uas-gofiber/middleware"
)

// RegisterUserRoutes mendaftarkan rute khusus Admin untuk manajemen user (FR-009)
func RegisterUserRoutes(v1 fiber.Router, userService *service.AdminManageUsersService) {
	
	const managePerm = "user:manage" 
	
	// Group /api/v1/users (Dilindungi AuthRequired)
	userRoute := v1.Group("/users", middleware.AuthRequired) 
	
	// FR-009 Flow 1: CREATE User
	// POST /api/v1/users
	userRoute.Post("/", middleware.RBACRequired(managePerm), userService.AdminCreateUser) 
	
	// FR-009 Flow 1: LIST All Users
	// GET /api/v1/users
	userRoute.Get("/", middleware.RBACRequired(managePerm), userService.AdminListUsers) 
	
	// FR-009 Flow 1: GET User by ID
	// GET /api/v1/users/:id
	userRoute.Get("/:id", middleware.RBACRequired(managePerm), userService.AdminGetUserByID)

	// FR-009 Flow 1: UPDATE User Data
	// PUT /api/v1/users/:id
	userRoute.Put("/:id", middleware.RBACRequired(managePerm), userService.AdminUpdateUser)
	
	// FR-009 Flow 1: DELETE User (Soft Delete)
	// DELETE /api/v1/users/:id
	userRoute.Delete("/:id", middleware.RBACRequired(managePerm), userService.AdminDeleteUser)
	
	// FR-009 Flow 2: Assign Role
	// PUT /api/v1/users/:id/role
	userRoute.Put("/:id/role", middleware.RBACRequired(managePerm), userService.AdminSetUserRole)
	
	// FR-009 Flow 4: Set Advisor
	// PUT /api/v1/users/students/:id/advisor
	userRoute.Put("/students/:id/advisor", middleware.RBACRequired(managePerm), userService.AdminSetAdvisor) 
}