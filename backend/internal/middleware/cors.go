package middleware

import (
	"strings"
	"vote/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CORS(cfg *config.Config) fiber.Handler {
	origins := cfg.AllowedOrigins
	if cfg.Environment == "development" {
		origins += ",http://localhost:3000,http://localhost:8080,http://127.0.0.1:3000,http://127.0.0.1:8080"
	}

	return cors.New(cors.Config{
		AllowOrigins:     strings.ReplaceAll(origins, " ", ""),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Device-Fingerprint",
		AllowCredentials: true,
	})
}
