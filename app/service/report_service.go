package service

import (
	"context"
	"fmt"
	"time"
	"github.com/gofiber/fiber/v2"
	
	repoMongo "github.com/safrizal-hk/uas-gofiber/app/repository/mongo"
	repoPostgres "github.com/safrizal-hk/uas-gofiber/app/repository/postgre"
	"github.com/safrizal-hk/uas-gofiber/middleware"
)

type ReportService struct {
	MongoRepo repoMongo.ReportMongoRepository
	PgRepo    repoPostgres.ReportPGRepository
}

func NewReportService(mongoRepo repoMongo.ReportMongoRepository, pgRepo repoPostgres.ReportPGRepository) *ReportService {
	return &ReportService{
		MongoRepo: mongoRepo,
		PgRepo:    pgRepo,
	}
}

func (s *ReportService) GetStatistics(c *fiber.Ctx) error {
	profile := middleware.GetUserProfileFromContext(c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var studentIDs []string

	switch profile.Role {
	case "Admin":
	case "Mahasiswa":
		studentID, err := s.PgRepo.FindStudentIdByUserID(profile.ID)
		if err != nil || studentID == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akun tidak terhubung ke data Mahasiswa.", "code": "403"})
		}
		studentIDs = append(studentIDs, studentID)
	case "Dosen Wali":
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "Fitur statistik Dosen Wali belum tersedia.", "code": "501"})
	default:
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Role tidak diizinkan mengakses statistik.", "code": "403"})
	}

	mongoStats, err := s.MongoRepo.GetAchievementStatistics(ctx, studentIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menjalankan agregasi MongoDB.", "code": "500"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"message": fmt.Sprintf("Statistik dimuat untuk role: %s.", profile.Role),
		"data": fiber.Map{
			"totalByTypeAndLevel": mongoStats,
		},
	})
}

func (s *ReportService) GetStudentReport(c *fiber.Ctx) error {
	profile := middleware.GetUserProfileFromContext(c)
	studentID := c.Params("id") // UUID Mahasiswa yang diminta

	if profile.Role == "Mahasiswa" {
		myStudentID, _ := s.PgRepo.FindStudentIdByUserID(profile.ID)
		if myStudentID != studentID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Anda tidak berhak melihat laporan ini.", "code": "403"})
		}
	}

	references, err := s.PgRepo.GetStudentAchievementReferences(studentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil data referensi prestasi.", "code": "500"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"message": "Laporan mahasiswa berhasil dimuat.",
		"total_prestasi": len(references),
		"data": references, 
	})
}