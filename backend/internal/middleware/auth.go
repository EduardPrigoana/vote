package middleware

import (
	"strings"
	"vote/internal/models"
	"vote/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func AuthRequired(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error: "Missing authorization header",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error: "Invalid authorization format",
			})
		}

		claims, err := utils.ValidateJWT(tokenString, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error: "Invalid or expired token",
			})
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

func AdminRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != "admin" && role != "superuser" {
			return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse{
				Error: "Admin access required",
			})
		}
		return c.Next()
	}
}

func SuperuserRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != "superuser" {
			return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse{
				Error: "Superuser access required",
			})
		}
		return c.Next()
	}
}
