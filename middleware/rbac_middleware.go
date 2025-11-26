package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func RBACRequired(requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userProfile := GetUserProfileFromContext(c)

		hasPermission := false
		
		for _, perm := range userProfile.Permissions {
			if perm == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Insufficient permissions (Need: " + requiredPermission + ")",
				"code": "403",
			})
		}

		return c.Next()
	}
}