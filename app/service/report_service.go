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
        // Admin: studentIDs kosong = Ambil Semua Data di Mongo (tidak difilter by studentId)
        studentIDs = []string{} 

    case "Mahasiswa":
        studentID, err := s.PgRepo.FindStudentIdByUserID(profile.ID)
        if err != nil || studentID == "" {
            return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akun tidak terhubung ke data Mahasiswa.", "code": "403"})
        }
        studentIDs = append(studentIDs, studentID)

    case "Dosen Wali":
        // --- [LOGIKA BARU DOSEN WALI] ---
        // 1. Cari ID Dosen
        lecturerID, err := s.PgRepo.FindLecturerIdByUserID(profile.ID)
        if err != nil || lecturerID == "" {
             return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akun tidak terhubung ke data Dosen Wali.", "code": "403"})
        }

        // 2. Ambil list ID mahasiswa bimbingan
        adviseeIDs, err := s.PgRepo.GetAdviseeStudentIDs(lecturerID)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil data bimbingan.", "code": "500"})
        }

        // Jika tidak punya bimbingan, kembalikan data kosong (bukan error)
        if len(adviseeIDs) == 0 {
             return c.Status(fiber.StatusOK).JSON(fiber.Map{
                "status": "success",
                "message": "Belum ada mahasiswa bimbingan.",
                "data": fiber.Map{"totalByTypeAndLevel": []interface{}{}},
            })
        }
        studentIDs = adviseeIDs

    default:
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Role tidak diizinkan mengakses statistik.", "code": "403"})
    }

    // Panggil Mongo Agregasi (Sama untuk semua role, bedanya cuma di filter studentIDs)
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