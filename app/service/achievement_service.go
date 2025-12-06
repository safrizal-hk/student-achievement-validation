package service

import (
	"context"
	"encoding/json"
	"errors"
	// "fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	modelMongo "github.com/safrizal-hk/uas-gofiber/app/model/mongo"
	modelPostgres "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
	repoMongo "github.com/safrizal-hk/uas-gofiber/app/repository/mongo"
	repoPostgres "github.com/safrizal-hk/uas-gofiber/app/repository/postgre"
	"github.com/safrizal-hk/uas-gofiber/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementService struct {
	MongoRepo repoMongo.AchievementMongoRepository
	PgRepo    repoPostgres.AchievementPGRepository
}

func NewAchievementService(mongoRepo repoMongo.AchievementMongoRepository, pgRepo repoPostgres.AchievementPGRepository) *AchievementService {
	return &AchievementService{
		MongoRepo: mongoRepo,
		PgRepo:    pgRepo,
	}
}

func (s *AchievementService) ListAllAchievements(c *fiber.Ctx) error {
	profile := middleware.GetUserProfileFromContext(c)
	var references []modelPostgres.AchievementReference
	var err error

	switch profile.Role {
	case "Admin":
		references, err = s.PgRepo.GetAllAchievementReferences()
	case "Mahasiswa":
		studentID, errLookup := s.PgRepo.FindStudentIdByUserID(profile.ID)
		if errLookup != nil || studentID == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akun tidak terhubung ke data Mahasiswa.", "code": "403"})
		}
		references, err = s.PgRepo.GetMyAchievements(studentID)
	case "Dosen Wali":
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
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Role tidak dikenal.", "code": "403"})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil data prestasi.", "code": "500"})
	}

	var mongoIDs []primitive.ObjectID
	for _, ref := range references {
		if oid, err := primitive.ObjectIDFromHex(ref.MongoAchievementID); err == nil {
			mongoIDs = append(mongoIDs, oid)
		}
	}

	details, err := s.MongoRepo.GetDetailsByIDs(context.Background(), mongoIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil detail dari MongoDB.", "code": "500"})
	}

	mongoDetailMap := make(map[string]modelMongo.AchievementMongo)
	for _, detail := range details {
		mongoDetailMap[detail.ID.Hex()] = detail
	}

	var finalData []fiber.Map
	for _, ref := range references {
		if detail, ok := mongoDetailMap[ref.MongoAchievementID]; ok {
			combined := fiber.Map{
				"id":              ref.ID,
				"student_id":      ref.StudentID,
				"status":          ref.Status,
				"submitted_at":    ref.SubmittedAt,
				"verified_at":     ref.VerifiedAt,
				"verified_by":     ref.VerifiedBy,
				"rejection_note":  ref.RejectionNote,
				"title":           detail.Title,
				"achievementType": detail.AchievementType,
				"description":     detail.Description,
				"points":          detail.Points,
				"tags":            detail.Tags,
				"details":         detail.Details,
				"attachments":     detail.Attachments,
				"created_at":      ref.CreatedAt,
			}
			finalData = append(finalData, combined)
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"total":  len(finalData),
		"data":   finalData,
	})
}

func (s *AchievementService) GetAchievementDetail(c *fiber.Ctx) error {
	achievementID := c.Params("id")

	ref, err := s.PgRepo.GetReferenceByID(achievementID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Prestasi tidak ditemukan", "code": "404"})
	}

	mongoID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	detail, err := s.MongoRepo.GetDetailByID(context.Background(), mongoID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil detail data", "code": "500"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"reference": ref,
			"detail":    detail,
		},
	})
}

func (s *AchievementService) SubmitPrestasi(c *fiber.Ctx) error {
	profile := middleware.GetUserProfileFromContext(c)
	if profile.Role != "Mahasiswa" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Hanya Mahasiswa yang dapat submit prestasi", "code": "403"})
	}

	studentID, err := s.PgRepo.FindStudentIdByUserID(profile.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Kesalahan server saat mencari data.", "code": "500"})
	}
	if studentID == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akun tidak terhubung ke data Mahasiswa.", "code": "403"})
	}

	req := new(modelMongo.AchievementInput)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input prestasi tidak valid", "code": "400"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mongoAch := modelMongo.AchievementMongo{
		StudentID:       studentID,
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Attachments:     req.Attachments,
		Tags:            req.Tags,
		Points:          req.Points,
	}
	createdMongo, err := s.MongoRepo.Create(ctx, &mongoAch)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menyimpan detail ke MongoDB", "code": "500"})
	}

	pgRef := modelPostgres.AchievementReference{
		StudentID:          studentID,
		MongoAchievementID: createdMongo.ID.Hex(),
	}
	createdRef, err := s.PgRepo.CreateReference(&pgRef)

	if err != nil {
		s.MongoRepo.DeleteByID(ctx, createdMongo.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menyimpan referensi ke PostgreSQL", "code": "500"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":   "success",
		"message":  "Prestasi berhasil disimpan sebagai DRAFT",
		"id":       createdRef.ID,
		"mongo_id": createdMongo.ID.Hex(),
	})
}

func (s *AchievementService) UpdatePrestasi(c *fiber.Ctx) error {
	profile := middleware.GetUserProfileFromContext(c)
	achievementID := c.Params("id")

	if profile.Role != "Mahasiswa" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak", "code": "403"})
	}

	ref, err := s.PgRepo.GetReferenceByID(achievementID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Prestasi tidak ditemukan", "code": "404"})
	}

	studentID, _ := s.PgRepo.FindStudentIdByUserID(profile.ID)
	if ref.StudentID != studentID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Bukan milik Anda", "code": "403"})
	}

	if ref.Status != modelPostgres.StatusDraft && ref.Status != modelPostgres.StatusRejected {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Tidak dapat mengubah prestasi yang sudah disubmit/diverifikasi", "code": "400"})
	}

	req := new(modelMongo.AchievementInput)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid", "code": "400"})
	}

	mongoID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	err = s.MongoRepo.Update(context.Background(), mongoID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal update data", "code": "500"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Prestasi berhasil diperbarui"})
}

func (s *AchievementService) DeletePrestasi(c *fiber.Ctx) error {
	profile := middleware.GetUserProfileFromContext(c)
	achievementID := c.Params("id")

	if profile.Role != "Mahasiswa" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Hanya Mahasiswa yang dapat menghapus", "code": "403"})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "ID Mongo corrupt", "code": "500"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = s.MongoRepo.SoftDelete(ctx, mongoObjectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal hapus detail Mongo", "code": "500"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success", "message": "Prestasi berhasil dihapus (soft deleted).",
	})
}

func (s *AchievementService) SubmitForVerification(c *fiber.Ctx) error {
	profile := middleware.GetUserProfileFromContext(c)
	achievementID := c.Params("id")

	if profile.Role != "Mahasiswa" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak", "code": "403"})
	}

	ref, err := s.PgRepo.GetReferenceByID(achievementID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Prestasi tidak ditemukan", "code": "404"})
	}

	studentID, err := s.PgRepo.FindStudentIdByUserID(profile.ID)
	if err != nil || studentID == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Data mahasiswa invalid", "code": "403"})
	}

	if ref.StudentID != studentID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Bukan milik Anda", "code": "403"})
	}

	updatedRef, err := s.PgRepo.UpdateStatusToSubmitted(achievementID)
	if err != nil {
		if errors.Is(err, errors.New("prestasi tidak ditemukan atau status sudah berubah")) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Gagal submit: Prestasi harus berstatus DRAFT.", "code": "400"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error(), "code": "500"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success", "message": "Prestasi berhasil disubmit untuk verifikasi", "new_status": updatedRef.Status,
	})
}

func (s *AchievementService) VerifyPrestasi(c *fiber.Ctx) error {
	profile := middleware.GetUserProfileFromContext(c)
	achievementID := c.Params("id")

	if profile.Role != "Dosen Wali" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak. Hanya Dosen Wali.", "code": "403"})
	}

	updatedRef, err := s.PgRepo.VerifyAchievement(achievementID, profile.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error(), "code": "400"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success", "message": "Prestasi berhasil diverifikasi.", "new_status": updatedRef.Status,
	})
}

func (s *AchievementService) RejectPrestasi(c *fiber.Ctx) error {
	profile := middleware.GetUserProfileFromContext(c)
	achievementID := c.Params("id")

	if profile.Role != "Dosen Wali" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak", "code": "403"})
	}

	var input map[string]string
	if err := json.Unmarshal(c.Body(), &input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid", "code": "400"})
	}
	rejectionNote, ok := input["rejection_note"]
	if !ok || strings.TrimSpace(rejectionNote) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Catatan penolakan wajib diisi.", "code": "400"})
	}

	updatedRef, err := s.PgRepo.RejectAchievement(achievementID, profile.ID, rejectionNote)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error(), "code": "400"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success", "message": "Prestasi ditolak.", "new_status": updatedRef.Status,
	})
}

func (s *AchievementService) GetHistory(c *fiber.Ctx) error {
	achievementID := c.Params("id")
	ref, err := s.PgRepo.GetReferenceByID(achievementID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Not Found"})
	}

	history := []fiber.Map{{"status": "created", "timestamp": ref.CreatedAt}}
	if ref.SubmittedAt != nil {
		history = append(history, fiber.Map{"status": "submitted", "timestamp": ref.SubmittedAt})
	}
	if ref.VerifiedAt != nil {
		history = append(history, fiber.Map{"status": string(ref.Status), "timestamp": ref.VerifiedAt, "note": ref.RejectionNote})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": history})
}

func (s *AchievementService) AddAttachment(c *fiber.Ctx) error {
	profile := middleware.GetUserProfileFromContext(c)
	achievementID := c.Params("id")

	if profile.Role != "Mahasiswa" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak. Hanya Mahasiswa.", "code": "403"})
	}

	ref, err := s.PgRepo.GetReferenceByID(achievementID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Prestasi tidak ditemukan", "code": "404"})
	}

	studentID, err := s.PgRepo.FindStudentIdByUserID(profile.ID)
	if err != nil || studentID == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Data mahasiswa tidak ditemukan.", "code": "403"})
	}
	
	if ref.StudentID != studentID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Anda tidak memiliki akses ke prestasi ini.", "code": "403"})
	}

	if ref.Status != modelPostgres.StatusDraft && ref.Status != modelPostgres.StatusRejected {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Tidak dapat mengubah lampiran pada prestasi yang sudah disubmit/diverifikasi", "code": "400"})
	}

	var attachment modelMongo.Attachment
	if err := c.BodyParser(&attachment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Data lampiran tidak valid", "code": "400"})
	}
	attachment.UploadedAt = time.Now()

	mongoID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	err = s.MongoRepo.AddAttachment(context.Background(), mongoID, attachment)
	
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menyimpan lampiran ke database", "code": "500"})
	}
	
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Lampiran berhasil ditambahkan"})
}	