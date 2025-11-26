package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/safrizal-hk/uas-gofiber/app/service"
)

// RegisterAuthRoutes mendaftarkan rute otentikasi di bawah prefix /auth
func RegisterAuthRoutes(v1 fiber.Router, authService *service.AuthService) {
	
	// Definisikan Sub-Group /auth
	authRoute := v1.Group("/auth")
	
	// Langsung menunjuk ke fungsi service.Login
	authRoute.Post("/login", authService.Login) 
	
	// authRoute.Post("/refresh", authService.RefreshToken) 
	
	// authRoute.Post("/logout", authService.Logout) 
	
	// authRoute.Get("/profile", middleware.AuthRequired, authService.Profile)
}