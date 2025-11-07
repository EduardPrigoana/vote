package middleware

import (
	"strings"
	"vote/internal/config"
	"vote/internal/models"

	"github.com/gofiber/fiber/v2"
)

func DomainRestriction(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/ws" {
			return c.Next()
		}

		origin := c.Get("Origin")
		referer := c.Get("Referer")

		if cfg.Environment == "development" {
			if strings.Contains(origin, "localhost") || strings.Contains(referer, "localhost") || origin == "" {
				return c.Next()
			}
		}

		allowedDomains := strings.Split(cfg.AllowedOrigins, ",")
		for _, domain := range allowedDomains {
			domain = strings.TrimSpace(domain)
			if strings.Contains(origin, domain) || strings.Contains(referer, domain) {
				return c.Next()
			}
		}

		if origin == "" && referer == "" {
			return c.Next()
		}

		return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse{
			Error: "Access denied",
		})
	}
}
