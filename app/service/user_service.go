package service

import (
	"fmt"
	"strings"
	// "errors"
	
	"github.com/gofiber/fiber/v2"
	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
	repo_postgre "github.com/safrizal-hk/uas-gofiber/app/repository/postgre"
	"github.com/safrizal-hk/uas-gofiber/utils"
)

// Struct Service Baru
type AdminManageUsersService struct {
	UserRepo repo_postgre.AdminManageUsersRepository
}

func NewAdminManageUsersService(userRepo repo_postgre.AdminManageUsersRepository) *AdminManageUsersService {
	return &AdminManageUsersService{UserRepo: userRepo}
}

// ----------------------------------------------------------------------
// FR-009 CRUD User (Penamaan Baru)
// ----------------------------------------------------------------------

// AdminCreateUser (POST /users) (FR-009 Flow 1, 2, 3)
func (s *AdminManageUsersService) AdminCreateUser(c *fiber.Ctx) error {
	req := new(model_postgre.UserCreateRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid", "code": "400"})
	}

	// 1. Dapatkan Role ID (Flow 2)
	role, err := s.UserRepo.GetRoleByName(req.RoleName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Role tidak valid: " + err.Error(), "code": "400"})
	}

	// 2. Hashing Password (Flow 1)
	hashedPass, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal hashing password", "code": "500"})
	}

	// 3. Panggil Repository
	newUser, err := s.UserRepo.CreateUser(req, role.ID, hashedPass)
	if err != nil {
		// Pengecekan Error Duplikasi (PostgreSQL specific error)
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "Username atau Email sudah terdaftar", "code": "409"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal membuat user", "detail": err.Error(), "code": "500"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status": "success",
		"message": "User berhasil dibuat",
		"id": newUser.ID,
		"role": req.RoleName,
	})
}

// AdminListUsers (GET /users)
func (s *AdminManageUsersService) AdminListUsers(c *fiber.Ctx) error {
	users, err := s.UserRepo.ListAllUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil daftar user", "code": "500"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"total": len(users),
		"data": users,
	})
}

// AdminGetUserByID (GET /users/:id)
func (s *AdminManageUsersService) AdminGetUserByID(c *fiber.Ctx) error {
	userID := c.Params("id")
	user, err := s.UserRepo.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": err.Error(), "code": "404"})
	}
	
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": user,
	})
}

// AdminUpdateUser (PUT /users/:id)
func (s *AdminManageUsersService) AdminUpdateUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	req := new(model_postgre.UserUpdateRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid", "code": "400"})
	}
	
	err := s.UserRepo.UpdateUser(userID, req)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": err.Error(), "code": "404"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"message": fmt.Sprintf("User %s berhasil diupdate.", userID),
	})
}

// AdminDeleteUser (DELETE /users/:id) (Soft Delete)
func (s *AdminManageUsersService) AdminDeleteUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	err := s.UserRepo.DeleteUser(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": err.Error(), "code": "404"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"message": fmt.Sprintf("User %s berhasil di-nonaktifkan (soft delete).", userID),
	})
}

// AdminSetUserRole (PUT /users/:id/role) (FR-009 Flow 2)
func (s *AdminManageUsersService) AdminSetUserRole(c *fiber.Ctx) error {
	userID := c.Params("id")
	req := new(model_postgre.AssignRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid", "code": "400"})
	}

	// 1. Dapatkan Role ID
	role, err := s.UserRepo.GetRoleByName(req.RoleName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Role tidak valid: " + err.Error(), "code": "400"})
	}

	// 2. Set Role di Repository
	err = s.UserRepo.SetUserRole(userID, role.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error(), "code": "400"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"message": fmt.Sprintf("Role user %s berhasil diubah menjadi %s.", userID, req.RoleName),
	})
}

// AdminSetAdvisor (PUT /users/students/:id/advisor) (FR-009 Flow 4)
func (s *AdminManageUsersService) AdminSetAdvisor(c *fiber.Ctx) error {
	studentID := c.Params("id") // Ini adalah students.id (UUID)
	req := new(model_postgre.SetAdvisorRequest)
	
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid", "code": "400"})
	}
	
	// Panggil Repository untuk update advisor_id
	err := s.UserRepo.SetStudentAdvisor(studentID, req.AdvisorID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error(), "code": "400"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"message": fmt.Sprintf("Advisor untuk Mahasiswa %s berhasil di-assign.", studentID),
	})
}