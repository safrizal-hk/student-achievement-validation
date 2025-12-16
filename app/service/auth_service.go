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

// Login godoc
// @Summary      Login Pengguna
// @Description  Otentikasi user menggunakan email/username dan password untuk mendapatkan JWT Token.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body model_postgre.LoginRequest true "Login Payload"
// @Success      200  {object}  map[string]interface{} "Success Login"
// @Failure      400  {object}  map[string]interface{} "Bad Request"
// @Failure      401  {object}  map[string]interface{} "Unauthorized (Salah Password/Tidak Aktif)"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /auth/login [post]
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

// RefreshToken godoc
// @Summary      Refresh Access Token
// @Description  Mendapatkan Access Token baru menggunakan Refresh Token yang valid.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body model_postgre.RefreshTokenRequest true "Refresh Token Payload"
// @Success      200  {object}  map[string]interface{} "Token Baru"
// @Failure      400  {object}  map[string]interface{} "Bad Request"
// @Failure      401  {object}  map[string]interface{} "Token Expired/Invalid"
// @Router       /auth/refresh [post]
func (s *AuthService) RefreshToken(c *fiber.Ctx) error {
	req := new(model_postgre.RefreshTokenRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Refresh token wajib diisi", "code": "400"})
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	token, err := jwt.ParseWithClaims(req.RefreshToken, &utils.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Refresh token tidak valid atau kadaluarsa", "code": "401"})
	}

	claims, ok := token.Claims.(*utils.JWTClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Format token salah", "code": "401"})
	}

	user, roleName, err := s.AuthRepo.FindUserByID(claims.UserProfile.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal memverifikasi user", "code": "500"})
	}
	if user == nil || !user.IsActive {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "User tidak valid atau sudah tidak aktif", "code": "401"})
	}

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

// Logout godoc
// @Summary      Logout
// @Description  Melakukan logout (Client-side clear token).
// @Tags         Authentication
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /auth/logout [post]
func (s *AuthService) Logout(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Logout berhasil.",
	})
}

// Profile godoc
// @Summary      Get User Profile
// @Description  Mendapatkan informasi profil user yang sedang login.
// @Tags         Authentication
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{} "User Profile"
// @Failure      404  {object}  map[string]interface{} "Not Found"
// @Failure      500  {object}  map[string]interface{} "Internal Error"
// @Router       /auth/profile [get]
func (s *AuthService) Profile(c *fiber.Ctx) error {
	profileFromCtx := middleware.GetUserProfileFromContext(c)
	
	user, roleName, err := s.AuthRepo.FindUserByID(profileFromCtx.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil profil", "code": "500"})
	}
	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "User tidak ditemukan", "code": "404"})
	}

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