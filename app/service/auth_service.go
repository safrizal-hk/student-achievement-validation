package service

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	model_postgre "github.com/safrizal-hk/uas-gofiber/app/model/postgre"
	repo_postgre "github.com/safrizal-hk/uas-gofiber/app/repository/postgre"
	"github.com/safrizal-hk/uas-gofiber/middleware"
	"github.com/safrizal-hk/uas-gofiber/utils"
)

type AuthService struct {
	AuthRepo repo_postgre.AuthRepository
}

func NewAuthService(repo repo_postgre.AuthRepository) *AuthService {
	return &AuthService{AuthRepo: repo}
}

// 1. Login (FR-001)
func (s *AuthService) Login(c *fiber.Ctx) error {
	req := new(model_postgre.LoginRequest)
	
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

	// 6. Generate Profile & Token
	profile := model_postgre.UserProfile{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Role:        roleName, 
		Permissions: permissions,
	}

	token, err := utils.GenerateJWT(profile, time.Minute*15) 
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal generate token", "code": "500"})
	}

	refreshToken, err := utils.GenerateJWT(profile, time.Hour*24*7)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal generate refresh token", "code": "500"})
	}

	// 7. Response
	resp := model_postgre.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         profile,
	}
	
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   resp,
	})
}

func (s *AuthService) RefreshToken(c *fiber.Ctx) error {
	req := new(model_postgre.RefreshTokenRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Refresh token wajib diisi", "code": "400"})
	}

	// 1. Validasi Token Manual
	jwtSecret := os.Getenv("JWT_SECRET")
	token, err := jwt.ParseWithClaims(req.RefreshToken, &utils.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	// Cek validitas signature dan expiration
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Refresh token tidak valid atau kadaluarsa", "code": "401"})
	}

	// 2. Ambil Claims
	claims, ok := token.Claims.(*utils.JWTClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Format token salah", "code": "401"})
	}

	// 3. Cek apakah user masih ada/aktif di DB (Security Check)
	// Menggunakan FindUserByID yang harus ada di Repository
	user, roleName, err := s.AuthRepo.FindUserByID(claims.UserProfile.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal memverifikasi user", "code": "500"})
	}
	if user == nil || !user.IsActive {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "User tidak valid atau sudah tidak aktif", "code": "401"})
	}

	// 4. Generate Token Baru
	permissions, _ := s.AuthRepo.GetPermissionsByRoleID(user.RoleID)
	newProfile := model_postgre.UserProfile{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Role:        roleName,
		Permissions: permissions,
	}

	newToken, err := utils.GenerateJWT(newProfile, time.Minute*15)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal membuat token baru", "code": "500"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"message": "Token berhasil diperbarui",
		"data": fiber.Map{
			"token": newToken,
		},
	})
}

// 3. Logout
func (s *AuthService) Logout(c *fiber.Ctx) error {
	// Karena menggunakan Stateless JWT, logout dilakukan di sisi client (menghapus token).
	// Di sini kita hanya mengembalikan respons sukses.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Logout berhasil.",
	})
}

// 4. Get Profile
func (s *AuthService) Profile(c *fiber.Ctx) error {
	// 1. Ambil data dari Context (disimpan oleh middleware AuthRequired)
	profileFromCtx := middleware.GetUserProfileFromContext(c)
	
	// 2. Ambil data segar dari DB (Optional tapi disarankan)
	user, roleName, err := s.AuthRepo.FindUserByID(profileFromCtx.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil profil", "code": "500"})
	}
	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "User tidak ditemukan", "code": "404"})
	}

	// 3. Reconstruct profile dengan data terbaru
	permissions, _ := s.AuthRepo.GetPermissionsByRoleID(user.RoleID)
	
	freshProfile := model_postgre.UserProfile{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Role:        roleName,
		Permissions: permissions,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   freshProfile,
	})
}