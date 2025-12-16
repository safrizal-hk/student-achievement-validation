package service

import (
	"fmt"
	"strings"
	
	"github.com/gofiber/fiber/v2"
	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
	repo_postgre "github.com/safrizal-hk/uas-gofiber/app/repository/postgre"
	"github.com/safrizal-hk/uas-gofiber/utils"
)

type AdminManageUsersService struct {
	UserRepo repo_postgre.AdminManageUsersRepository
}

func NewAdminManageUsersService(userRepo repo_postgre.AdminManageUsersRepository) *AdminManageUsersService {
	return &AdminManageUsersService{UserRepo: userRepo}
}

// AdminCreateUser godoc
// @Summary      Buat User Baru (Admin)
// @Description  Admin membuat user baru secara manual dengan role tertentu.
// @Tags         Admin - User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body      model_postgre.UserCreateRequest true "Data User Baru"
// @Success      201  {object}  map[string]interface{} "User Created"
// @Failure      400  {object}  map[string]interface{} "Bad Request"
// @Failure      409  {object}  map[string]interface{} "Conflict (Email/Username exists)"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /users [post]
func (s *AdminManageUsersService) AdminCreateUser(c *fiber.Ctx) error {
	req := new(model_postgre.UserCreateRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid", "code": "400"})
	}

	role, err := s.UserRepo.GetRoleByName(req.RoleName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Role tidak valid: " + err.Error(), "code": "400"})
	}

	hashedPass, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal hashing password", "code": "500"})
	}

	newUser, err := s.UserRepo.CreateUser(req, role.ID, hashedPass)
	if err != nil {
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

// AdminListUsers godoc
// @Summary      List Semua User
// @Description  Melihat daftar semua user di sistem.
// @Tags         Admin - User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{} "List Users"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /users [get]
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

// AdminGetUserByID godoc
// @Summary      Detail User
// @Description  Melihat detail user berdasarkan ID.
// @Tags         Admin - User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  map[string]interface{} "User Detail"
// @Failure      404  {object}  map[string]interface{} "Not Found"
// @Router       /users/{id} [get]
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

// AdminUpdateUser godoc
// @Summary      Update User
// @Description  Mengubah data user.
// @Tags         Admin - User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID (UUID)"
// @Param        body body      model_postgre.UserUpdateRequest true "Data Update"
// @Success      200  {object}  map[string]interface{} "Success"
// @Failure      400  {object}  map[string]interface{} "Bad Request"
// @Failure      404  {object}  map[string]interface{} "Not Found"
// @Router       /users/{id} [put]
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

// AdminDeleteUser godoc
// @Summary      Hapus User (Soft Delete)
// @Description  Menonaktifkan user (Soft Delete).
// @Tags         Admin - User Management
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  map[string]interface{} "Success"
// @Failure      404  {object}  map[string]interface{} "Not Found"
// @Router       /users/{id} [delete]
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

// AdminSetUserRole godoc
// @Summary      Ubah Role User
// @Description  Mengganti role user (Misal: dari Mahasiswa ke Admin).
// @Tags         Admin - User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID (UUID)"
// @Param        body body      model_postgre.AssignRoleRequest true "Role Name"
// @Success      200  {object}  map[string]interface{} "Success"
// @Failure      400  {object}  map[string]interface{} "Bad Request"
// @Router       /users/{id}/role [put]
func (s *AdminManageUsersService) AdminSetUserRole(c *fiber.Ctx) error {
	userID := c.Params("id")
	req := new(model_postgre.AssignRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid", "code": "400"})
	}

	role, err := s.UserRepo.GetRoleByName(req.RoleName)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Role tidak valid: " + err.Error(), "code": "400"})
	}

	err = s.UserRepo.SetUserRole(userID, role.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error(), "code": "400"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"message": fmt.Sprintf("Role user %s berhasil diubah menjadi %s.", userID, req.RoleName),
	})
}

// AdminSetAdvisor godoc
// @Summary      Assign Dosen Wali
// @Description  Menghubungkan Mahasiswa dengan Dosen Wali.
// @Tags         Admin - User Management
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Student ID (UUID)"
// @Param        body body      model_postgre.SetAdvisorRequest true "Advisor ID"
// @Success      200  {object}  map[string]interface{} "Success"
// @Failure      400  {object}  map[string]interface{} "Bad Request"
// @Router       /students/{id}/advisor [put]
func (s *AdminManageUsersService) AdminSetAdvisor(c *fiber.Ctx) error {
	studentID := c.Params("id")
	req := new(model_postgre.SetAdvisorRequest)
	
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid", "code": "400"})
	}
	
	err := s.UserRepo.SetStudentAdvisor(studentID, req.AdvisorID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error(), "code": "400"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"message": fmt.Sprintf("Advisor untuk Mahasiswa %s berhasil di-assign.", studentID),
	})
}