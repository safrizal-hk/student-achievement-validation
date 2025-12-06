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

	authRepo := repo_postgre.NewAuthRepository(dbConn.PgDB) 
	achievementPgRepo := repo_postgre.NewAchievementPGRepository(dbConn.PgDB)
	achievementMongoRepo := repo_mongo.NewAchievementMongoRepository(dbConn.MongoDB)
	reportPgRepo := repo_postgre.NewReportPGRepository(dbConn.PgDB)
	reportMongoRepo := repo_mongo.NewReportMongoRepository(dbConn.MongoDB)
	userRepo := repo_postgre.NewAdminManageUsersRepository(dbConn.PgDB)
	studentRepo := repo_postgre.NewStudentRepository(dbConn.PgDB)
	lecturerRepo := repo_postgre.NewLecturerRepository(dbConn.PgDB)

	authService := service.NewAuthService(authRepo) 
	achievementService := service.NewAchievementService(achievementMongoRepo, achievementPgRepo) 
	userService := service.NewAdminManageUsersService(userRepo) 
	reportService := service.NewReportService(reportMongoRepo, reportPgRepo)
	studentService := service.NewStudentService(studentRepo, achievementPgRepo) 
	lecturerService := service.NewLecturerService(lecturerRepo)
	
	RegisterAuthRoutes(v1, authService) 
	RegisterAchievementRoutes(v1, achievementService)
	RegisterUserRoutes(v1, userService) 
	RegisterReportRoutes(v1, reportService)
	RegisterStudentRoutes(v1, studentService)
	RegisterLecturerRoutes(v1, lecturerService)
}