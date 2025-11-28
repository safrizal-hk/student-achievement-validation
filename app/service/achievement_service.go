package service

import (
	"context"
	"errors"
	"time"
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	modelMongo "github.com/safrizal-hk/uas-gofiber/app/model/mongo"
	modelPostgres "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
	repoMongo "github.com/safrizal-hk/uas-gofiber/app/repository/mongo"
	repoPostgres "github.com/safrizal-hk/uas-gofiber/app/repository/postgre"
	"github.com/safrizal-hk/uas-gofiber/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AchievementService Struct - Tetap di package 'service'
type AchievementService struct {
	MongoRepo repoMongo.AchievementMongoRepository
	PgRepo repoPostgres.AchievementPGRepository
}

func NewAchievementService(mongoRepo repoMongo.AchievementMongoRepository, pgRepo repoPostgres.AchievementPGRepository) *AchievementService {
	return &AchievementService{
		MongoRepo: mongoRepo,
		PgRepo: pgRepo,
	}
}

// ----------------------------------------------------------------------
// FR-003, FR-004, FR-005 (Submit, Verification, Delete) - Tetap sama
// ----------------------------------------------------------------------

// SubmitPrestasi (FR-003) - Logic tidak berubah
func (s *AchievementService) SubmitPrestasi(c *fiber.Ctx) error {
	// ... (Logic FR-003: Auth, Lookup studentID, Write Mongo, Write PG - TIDAK ADA ROLLBACK) ...
	profile := middleware.GetUserProfileFromContext(c)
	if profile.Role != "Mahasiswa" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Hanya Mahasiswa yang dapat submit prestasi", "code": "403"})
	}
	studentID, err := s.PgRepo.FindStudentIdByUserID(profile.ID)
	if err != nil || studentID == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akun tidak terhubung ke data Mahasiswa.", "code": "403"})
	}
	req := new(modelMongo.AchievementInput)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input prestasi tidak valid", "code": "400"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mongoAch := modelMongo.AchievementMongo{
		StudentID: studentID, AchievementType: req.AchievementType, Title: req.Title, Description: req.Description, 
		Details: req.Details, Attachments: req.Attachments, Tags: req.Tags, Points: req.Points,
	}
	createdMongo, err := s.MongoRepo.Create(ctx, &mongoAch)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menyimpan detail ke MongoDB", "code": "500"})
	}
	pgRef := modelPostgres.AchievementReference{
		StudentID: studentID, MongoAchievementID: createdMongo.ID.Hex(),
	}
	createdRef, err := s.PgRepo.CreateReference(&pgRef)
	
	if err != nil {
		// Rollback logic (JIKA DIIMPLEMENTASI)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menyimpan referensi ke PostgreSQL", "code": "500"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status": "success", "message": "Prestasi berhasil disimpan sebagai DRAFT", "id": createdRef.ID, "mongo_id": createdMongo.ID.Hex(),
	})
}

// SubmitForVerification (FR-004) - Logic tidak berubah
func (s *AchievementService) SubmitForVerification(c *fiber.Ctx) error {
	// ... (Logic FR-004: Update status ke Submitted) ...
	profile := middleware.GetUserProfileFromContext(c)
	achievementID := c.Params("id")

	if profile.Role != "Mahasiswa" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Hanya Mahasiswa yang dapat submit untuk verifikasi", "code": "403"})
	}
	
	updatedRef, err := s.PgRepo.UpdateStatusToSubmitted(achievementID)
	
	if err != nil {
		if errors.Is(err, errors.New("prestasi tidak ditemukan atau status sudah berubah")) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": err.Error(), "code": "404"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error(), "code": "500"})
	}
	
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success", "message": "Prestasi berhasil disubmit untuk verifikasi", "new_status": updatedRef.Status,
	})
}

// DeletePrestasi (FR-005) - Logic tidak berubah
func (s *AchievementService) DeletePrestasi(c *fiber.Ctx) error {
	// ... (Logic FR-005: Soft delete di PG dan Mongo) ...
	profile := middleware.GetUserProfileFromContext(c)
	achievementID := c.Params("id") 

	if profile.Role != "Mahasiswa" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Hanya Mahasiswa yang dapat menghapus prestasi", "code": "403"})
	}

	studentID, err := s.PgRepo.FindStudentIdByUserID(profile.ID)
	if err != nil || studentID == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akun tidak terhubung ke data Mahasiswa.", "code": "403"})
	}

	deletedRef, err := s.PgRepo.SoftDeleteReference(achievementID, studentID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error(), "code": "400"})
	}

	mongoObjectID, err := primitive.ObjectIDFromHex(deletedRef.MongoAchievementID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengonversi Mongo ID.", "code": "500"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = s.MongoRepo.SoftDelete(ctx, mongoObjectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menghapus detail di MongoDB.", "code": "500"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success", "message": "Prestasi berhasil dihapus (soft deleted).",
	})
}

// ----------------------------------------------------------------------
// FR-006 & FR-010: View Achievements (Filtering Logic)
// ----------------------------------------------------------------------

// ListAllAchievements menangani GET /api/v1/achievements dengan filter role
func (s *AchievementService) ListAllAchievements(c *fiber.Ctx) error {
	profile := middleware.GetUserProfileFromContext(c)
	
	var references []modelPostgres.AchievementReference
	var err error
	
	// 1. Tentukan query berdasarkan Role
	switch profile.Role {
	case "Admin":
		// FR-010: Admin dapat melihat semua prestasi
		references, err = s.PgRepo.GetAllAchievementReferences()
		
	case "Mahasiswa":
		// Mahasiswa: Hanya melihat prestasi miliknya sendiri
		studentID, errLookup := s.PgRepo.FindStudentIdByUserID(profile.ID)
		if errLookup != nil || studentID == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akun tidak terhubung ke data Mahasiswa.", "code": "403"})
		}
		references, err = s.PgRepo.GetMyAchievements(studentID)
		
	case "Dosen Wali":
		// FR-006: Dosen wali melihat prestasi mahasiswa bimbingannya
		lecturerID, errLookup := s.PgRepo.FindLecturerIdByUserID(profile.ID)
		if errLookup != nil || lecturerID == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akun tidak terhubung ke data Dosen Wali.", "code": "403"})
		}
		
		studentIDs, errLookup := s.PgRepo.GetAdviseeStudentIDs(lecturerID)
		if errLookup != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil daftar mahasiswa bimbingan.", "code": "500"})
		}
		
		if len(studentIDs) > 0 {
			references, err = s.PgRepo.GetAchievementsByStudentIDs(studentIDs)
		} else {
			references = []modelPostgres.AchievementReference{}
		}
		
	default:
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Role tidak dikenal atau tidak diizinkan.", "code": "403"})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil data prestasi.", "code": "500"})
	}
	
	// 2. Kumpulkan mongo_achievement_id
	var mongoIDs []primitive.ObjectID
	for _, ref := range references {
		objID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
		mongoIDs = append(mongoIDs, objID)
	}

	// 3. Fetch detail dari MongoDB (FR-006 Flow 3)
	details, err := s.MongoRepo.GetDetailByMongoIDs(context.Background(), mongoIDs) 
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil detail prestasi dari MongoDB.", "code": "500"})
	}

	// 4. Gabungkan Data (Merge)
	mongoDetailMap := make(map[string]modelMongo.AchievementMongo)
	for _, detail := range details {
		mongoDetailMap[detail.ID.Hex()] = detail
	}

	var finalData []interface{} 
	for _, ref := range references {
		mongoIDHex := ref.MongoAchievementID
		if detail, found := mongoDetailMap[mongoIDHex]; found {
			combined := fiber.Map{
				"id": ref.ID,
				"status": ref.Status,
				"submittedAt": ref.SubmittedAt,
				"verifiedAt": ref.VerifiedAt,
				"title": detail.Title,
				"achievementType": detail.AchievementType,
				"details": detail.Details,
				"tags": detail.Tags,
				// ... dan semua field lain yang relevan
			}
			finalData = append(finalData, combined)
		}
	}
	
	// 5. Return Success
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"message": "Data prestasi berhasil dimuat.",
		"total": len(finalData),
		"data": finalData, 
	})
}

// VerifyPrestasi menangani alur FR-007
func (s *AchievementService) VerifyPrestasi(c *fiber.Ctx) error {
    profile := middleware.GetUserProfileFromContext(c)
    achievementID := c.Params("id") // Ambil ID Prestasi dari URL
    
    // 1. Validasi Role (Actor: Dosen Wali)
    if profile.Role != "Dosen Wali" {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak. Hanya Dosen Wali.", "code": "403"})
    }
    
    // TODO: [PENTING] SECURITY CHECK:
    // Pastikan Dosen Wali ini adalah pembimbing (advisor) dari mahasiswa pemilik prestasi.
    
    // 2. Panggil Repository (Flow 3 & 4: Update status, set verified_by/at)
    updatedRef, err := s.PgRepo.VerifyAchievement(achievementID, profile.ID)
    
    if err != nil {
        // Error handler (Not Found/Precondition failed)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error(), "code": "400"})
    }

    // 3. Return updated status (Flow 5)
    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "status": "success",
        "message": "Prestasi berhasil diverifikasi.",
        "new_status": updatedRef.Status, // Akan bernilai 'verified'
        "verified_at": updatedRef.VerifiedAt,
    })
}

// RejectPrestasi menangani alur FR-008
func (s *AchievementService) RejectPrestasi(c *fiber.Ctx) error {
    profile := middleware.GetUserProfileFromContext(c)
    achievementID := c.Params("id") // Ambil ID Prestasi dari URL
    
    // 1. Validasi Role (Actor: Dosen Wali)
    if profile.Role != "Dosen Wali" {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak. Hanya Dosen Wali.", "code": "403"})
    }

    // 2. Dapatkan Input Rejection Note dari Body JSON (Flow 1)
    var input map[string]string
    
    // Unmarshal ke map untuk menghindari struct baru
    if err := json.Unmarshal(c.Body(), &input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid (body harus berupa JSON).", "code": "400"})
    }

    rejectionNote, ok := input["rejection_note"]

    // 3. Cek Precondition (Catatan Wajib Diisi)
    if !ok || strings.TrimSpace(rejectionNote) == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Catatan penolakan (rejection_note) wajib diisi.", "code": "400"})
    }
    
    // TODO: SECURITY CHECK: Pastikan Dosen Wali adalah Advisor dari Mahasiswa pemilik prestasi (Penting!)
    
    // 4. Panggil Repository (Flow 2 & 3)
    updatedRef, err := s.PgRepo.RejectAchievement(achievementID, profile.ID, rejectionNote)
    
    if err != nil {
        // Error handler (Not Found/Precondition failed)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error(), "code": "400"})
    }

    // 5. Return updated status (Flow 5)
    // Flow 4 (Notification) diabaikan.
    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "status": "success",
        "message": "Prestasi berhasil ditolak.",
        "new_status": updatedRef.Status, 
        "rejection_note": updatedRef.RejectionNote,
    })
}