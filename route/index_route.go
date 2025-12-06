package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/safrizal-hk/uas-gofiber/app/service" 
	repo_postgre "github.com/safrizal-hk/uas-gofiber/app/repository/postgre" 
	repo_mongo "github.com/safrizal-hk/uas-gofiber/app/repository/mongo" 
	"github.com/safrizal-hk/uas-gofiber/config"
)

func RegisterAllRoutes(app *fiber.App, dbConn *config.Database) {
	
	v1 := app.Group("/api/v1") 

	// 1. Wiring Repositories
	authRepo := repo_postgre.NewAuthRepository(dbConn.PgDB) 
	achievementPgRepo := repo_postgre.NewAchievementPGRepository(dbConn.PgDB)
	achievementMongoRepo := repo_mongo.NewAchievementMongoRepository(dbConn.MongoDB)
	reportPgRepo := repo_postgre.NewReportPGRepository(dbConn.PgDB)
	reportMongoRepo := repo_mongo.NewReportMongoRepository(dbConn.MongoDB)
	
	// Wiring Repo Admin (FR-009)
	userRepo := repo_postgre.NewAdminManageUsersRepository(dbConn.PgDB) 

	// 2. Wiring Services
	authService := service.NewAuthService(authRepo) 
	achievementService := service.NewAchievementService(achievementMongoRepo, achievementPgRepo) 
	
	// Wiring Service Admin (FR-009)
	userService := service.NewAdminManageUsersService(userRepo) 
	reportService := service.NewReportService(reportMongoRepo, reportPgRepo) // Service Report
	
	// ---------- Pendaftaran Grup Rute ----------
	
	// A. Rute Otentikasi (Public)
	RegisterAuthRoutes(v1, authService) 
	
	// B. Rute Prestasi (Terproteksi)
	RegisterAchievementRoutes(v1, achievementService)
	
	// C. Rute User Management (Terproteksi - FR-009)
	RegisterUserRoutes(v1, userService) 

	RegisterReportRoutes(v1, reportService)
}