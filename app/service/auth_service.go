package service

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/safrizal-hk/uas-gofiber/app/model"
	"github.com/safrizal-hk/uas-gofiber/app/repository"
	"github.com/safrizal-hk/uas-gofiber/utils" 
)

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthService struct {
	AuthRepo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) *AuthService {
	return &AuthService{AuthRepo: repo}
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	req := new(LoginRequest)
	
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Input tidak valid", "code": "400"})
	}
	
	user, roleName, err := s.AuthRepo.FindUserByEmailOrUsername(req.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Terjadi kesalahan server", "code": "500"})
	}
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Kredensial tidak valid", "code": "401"})
	}

	if !user.IsActive {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Akun tidak aktif", "code": "401"})
	}

	cleanHash := strings.TrimSpace(user.PasswordHash) 
	
	if !utils.CheckPasswordHash(req.Password, cleanHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Kredensial tidak valid", "code": "401"})
	}

	permissions, err := s.AuthRepo.GetPermissionsByRoleID(user.RoleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil izin user", "code": "500"})
	}

	profile := model.UserProfile{
		ID: user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Role: roleName, 
		Permissions: permissions,
	}

	token, _ := utils.GenerateJWT(profile, time.Minute*15) 
	refreshToken, _ := utils.GenerateJWT(profile, time.Hour*24*7) 

	resp := model.LoginResponse{
		Token: token,
		RefreshToken: refreshToken,
		User: profile,
	}
	
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": resp,
	})
}