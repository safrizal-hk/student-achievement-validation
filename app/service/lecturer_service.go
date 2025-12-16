package service

import (
	"github.com/gofiber/fiber/v2"
	repo_postgre "github.com/safrizal-hk/uas-gofiber/app/repository/postgre"
)

type LecturerService struct {
	LecturerRepo repo_postgre.LecturerRepository
}

func NewLecturerService(lecturerRepo repo_postgre.LecturerRepository) *LecturerService {
	return &LecturerService{LecturerRepo: lecturerRepo}
}

// ListLecturers godoc
// @Summary      List Semua Dosen
// @Description  Mendapatkan daftar semua dosen yang terdaftar di sistem.
// @Tags         Lecturers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{} "List Dosen"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /lecturers [get]
func (s *LecturerService) ListLecturers(c *fiber.Ctx) error {
	lecturers, err := s.LecturerRepo.GetAllLecturers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil data dosen", "code": "500"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": lecturers})
}

// GetLecturerAdvisees godoc
// @Summary      List Mahasiswa Bimbingan
// @Description  Mendapatkan daftar mahasiswa yang dibimbing oleh Dosen tertentu berdasarkan ID Dosen.
// @Tags         Lecturers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Lecturer ID (UUID)"
// @Success      200  {object}  map[string]interface{} "List Mahasiswa Bimbingan"
// @Failure      404  {object}  map[string]interface{} "Dosen Tidak Ditemukan"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /lecturers/{id}/advisees [get]
func (s *LecturerService) GetLecturerAdvisees(c *fiber.Ctx) error {
	lecturerID := c.Params("id")

	_, err := s.LecturerRepo.GetLecturerByID(lecturerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Dosen tidak ditemukan", "code": "404"})
	}

	advisees, err := s.LecturerRepo.GetAdviseesByLecturerID(lecturerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil data bimbingan", "code": "500"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success", 
		"data": advisees,
		"total": len(advisees),
	})
}