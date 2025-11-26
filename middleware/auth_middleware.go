package middleware

import (
	"errors"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/safrizal-hk/uas-gofiber/app/model"
	"github.com/safrizal-hk/uas-gofiber/utils"
)

func AuthRequired(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Missing or invalid token", "code": "401"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Server configuration error", "code": "500"})
	}

	token, err := jwt.ParseWithClaims(tokenString, &utils.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid or expired token", "code": "401"})
	}

	claims, ok := token.Claims.(*utils.JWTClaims)
	if !ok || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid token claims", "code": "401"})
	}
	
	c.Locals("userProfile", claims.UserProfile)

	return c.Next()
}

func GetUserProfileFromContext(c *fiber.Ctx) model.UserProfile {
	if profile, ok := c.Locals("userProfile").(model.UserProfile); ok {
		return profile
	}
	return model.UserProfile{} 
}