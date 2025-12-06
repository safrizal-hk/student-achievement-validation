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

func (s *LecturerService) ListLecturers(c *fiber.Ctx) error {
	lecturers, err := s.LecturerRepo.GetAllLecturers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil data dosen", "code": "500"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": lecturers})
}

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