package service

import (
	"context"
	"fmt"
	"time"
	
	"github.com/gofiber/fiber/v2"
	// modelPostgres dihapus karena belum digunakan di logic saat ini (atau gunakan _ jika perlu init)
	
	repoMongo "github.com/safrizal-hk/uas-gofiber/app/repository/mongo"
	repoPostgres "github.com/safrizal-hk/uas-gofiber/app/repository/postgre"
	"github.com/safrizal-hk/uas-gofiber/middleware"
)

// ReportService Struct
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

// GetStatisticsHandler (GET /api/v1/reports/statistics) - FR-011
func (s *ReportService) GetStatisticsHandler(c *fiber.Ctx) error {
	profile := middleware.GetUserProfileFromContext(c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var studentIDs []string // Daftar Mahasiswa yang datanya akan diagregasi

	switch profile.Role {
	case "Admin":
		// Admin: Melihat semua data (studentIDs kosong, query MongoDB akan mengambil semua)
	case "Mahasiswa":
		// Mahasiswa: Hanya data sendiri
		studentID, err := s.PgRepo.FindStudentIdByUserID(profile.ID)
		if err != nil || studentID == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akun tidak terhubung ke data Mahasiswa.", "code": "403"})
		}
		studentIDs = append(studentIDs, studentID)
	case "Dosen Wali":
		// Dosen Wali: Data mahasiswa bimbingan 
		// TODO: Implementasi lookup dosen wali di repo report jika diperlukan di masa depan
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "Fitur statistik Dosen Wali belum tersedia.", "code": "501"})
	default:
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Role tidak diizinkan mengakses statistik.", "code": "403"})
	}

	// 1. Get statistik dari MongoDB (Total per Tipe, Distribusi Level)
	mongoStats, err := s.MongoRepo.GetAchievementStatistics(ctx, studentIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menjalankan agregasi MongoDB.", "code": "500"})
	}

	// 2. Gabungkan dan Format Hasil
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"message": fmt.Sprintf("Statistik dimuat untuk role: %s.", profile.Role),
		"data": fiber.Map{
			"totalByTypeAndLevel": mongoStats,
			// Top Students bisa ditambahkan di sini nanti
		},
	})
}

// GetStudentReportHandler (GET /api/v1/reports/student/:id)
func (s *ReportService) GetStudentReportHandler(c *fiber.Ctx) error {
	profile := middleware.GetUserProfileFromContext(c)
	studentID := c.Params("id") // UUID Mahasiswa yang diminta

	// 1. Otorisasi Akses (Security): Hanya Admin dan Dosen Wali (untuk bimbingan sendiri) yang boleh
	// Mahasiswa hanya boleh melihat report miliknya sendiri
	// Logic ini perlu disesuaikan: studentID adalah students.id (bukan users.id)
	
	// Untuk saat ini, kita izinkan Admin melihat semua.
	if profile.Role == "Mahasiswa" {
		// Perlu validasi apakah studentID param sama dengan studentID milik user login
		myStudentID, _ := s.PgRepo.FindStudentIdByUserID(profile.ID)
		if myStudentID != studentID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Anda tidak berhak melihat laporan ini.", "code": "403"})
		}
	}

	// 2. Ambil References Prestasi (PG)
	references, err := s.PgRepo.GetStudentAchievementReferences(studentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil data referensi prestasi.", "code": "500"})
	}

	// 3. Return Data
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"message": "Laporan mahasiswa berhasil dimuat.",
		"total_prestasi": len(references),
		"data": references, 
	})
}