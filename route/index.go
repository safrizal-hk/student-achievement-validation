package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/safrizal-hk/uas-gofiber/app/service" 
	"github.com/safrizal-hk/uas-gofiber/app/repository" 
	"github.com/safrizal-hk/uas-gofiber/config"
)

// RegisterAllRoutes adalah fungsi utama untuk mendaftarkan semua rute aplikasi
func RegisterAllRoutes(app *fiber.App, dbConn *config.Database) {
	
	// Definisikan Group API Utama
	v1 := app.Group("/api/v1") 

	// 1. Wiring Auth & Core Services (PostgreSQL)
	authRepo := repository.NewAuthRepository(dbConn.PgDB) 
	authService := service.NewAuthService(authRepo) 
	
	// 2. Wiring Achievement Services (Hybrid DB)
	// achievementPgRepo := repository.NewAchievementPGRepository(dbConn.PgDB)
	// achievementMongoRepo := repository.NewAchievementMongoRepository(dbConn.MongoDB)
	// achievementService := service.NewAchievementService(achievementPgRepo, achievementMongoRepo)

	
	// ---------- Pendaftaran Grup Rute ----------
	
	// A. Rute Otentikasi (Login, Refresh)
	RegisterAuthRoutes(v1, authService) 
	
	// B. Rute Prestasi (Memerlukan Auth Required & RBAC)
	// RegisterAchievementRoutes(v1, achievementService)
	
	// C. Rute Manajemen User (Hanya Admin)
	// RegisterUserRoutes(v1, userService)

	// ... daftarkan grup rute lainnya di sini
}