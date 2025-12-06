package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/safrizal-hk/uas-gofiber/app/service"
	"github.com/safrizal-hk/uas-gofiber/middleware" // Diperlukan untuk AuthRequired
)

func RegisterAuthRoutes(v1 fiber.Router, authService *service.AuthService) {
	authRoute := v1.Group("/auth")
	
	// Public Routes
	authRoute.Post("/login", authService.Login)
	authRoute.Post("/refresh", authService.RefreshToken) // Biasanya public, validasi token di body
	
	// Protected Routes (Butuh Login)
	authRoute.Post("/logout", middleware.AuthRequired, authService.Logout)
	authRoute.Get("/profile", middleware.AuthRequired, authService.Profile)
}