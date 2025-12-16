package service

import (
	"github.com/gofiber/fiber/v2"
	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
	repo_postgre "github.com/safrizal-hk/uas-gofiber/app/repository/postgre"
	// "github.com/safrizal-hk/uas-gofiber/middleware"
)

type StudentService struct {
	StudentRepo repo_postgre.StudentRepository
	AchievementRepo repo_postgre.AchievementPGRepository
}

func NewStudentService(studentRepo repo_postgre.StudentRepository, achievementRepo repo_postgre.AchievementPGRepository) *StudentService {
	return &StudentService{
		StudentRepo: studentRepo,
		AchievementRepo: achievementRepo,
	}
}

// ListStudents godoc
// @Summary      List Semua Mahasiswa
// @Description  Mendapatkan daftar semua mahasiswa yang terdaftar.
// @Tags         Students
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{} "List Mahasiswa"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /students [get]
func (s *StudentService) ListStudents(c *fiber.Ctx) error {
	students, err := s.StudentRepo.GetAllStudents()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil data mahasiswa", "code": "500"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": students})
}

// GetStudentDetail godoc
// @Summary      Detail Mahasiswa
// @Description  Mendapatkan detail profil mahasiswa berdasarkan ID.
// @Tags         Students
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Student ID (UUID)"
// @Success      200  {object}  map[string]interface{} "Detail Mahasiswa"
// @Failure      404  {object}  map[string]interface{} "Not Found"
// @Router       /students/{id} [get]
func (s *StudentService) GetStudentDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	student, err := s.StudentRepo.GetStudentDetail(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Mahasiswa tidak ditemukan", "code": "404"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": student})
}

// SetAdvisor godoc
// @Summary      Set Dosen Wali
// @Description  Menentukan atau mengubah dosen wali untuk mahasiswa tertentu.
// @Tags         Students
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string                         true "Student ID (UUID)"
// @Param        body body      model_postgre.SetAdvisorRequest true "Advisor ID Payload"
// @Success      200  {object}  map[string]interface{} "Success"
// @Failure      400  {object}  map[string]interface{} "Bad Request"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /students/{id}/advisor [put]
func (s *StudentService) SetAdvisor(c *fiber.Ctx) error {
	id := c.Params("id")
	req := new(model_postgre.SetAdvisorRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid", "code": "400"})
	}

	err := s.StudentRepo.SetStudentAdvisor(id, req.AdvisorID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal set advisor: " + err.Error(), "code": "500"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Advisor berhasil diatur"})
}

// GetStudentAchievements godoc
// @Summary      List Prestasi Mahasiswa
// @Description  Melihat daftar prestasi milik mahasiswa tertentu (Dilihat oleh Admin/Dosen).
// @Tags         Students
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Student ID (UUID)"
// @Success      200  {object}  map[string]interface{} "List Prestasi"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /students/{id}/achievements [get]
func (s *StudentService) GetStudentAchievements(c *fiber.Ctx) error {
	studentID := c.Params("id")
    
	achievements, err := s.AchievementRepo.GetAchievementsByStudentIDs([]string{studentID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil prestasi mahasiswa", "code": "500"})
	}
	
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": achievements})
}