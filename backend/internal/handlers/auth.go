package handlers

import (
	"database/sql"
	"time"
	"vote/internal/database"
	"vote/internal/models"
	"vote/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	DB        *database.Database
	JWTSecret string
	JWTExpiry int64
}

func NewAuthHandler(db *database.Database, jwtSecret string, jwtExpiry int64) *AuthHandler {
	return &AuthHandler{
		DB:        db,
		JWTSecret: jwtSecret,
		JWTExpiry: jwtExpiry,
	}
}

// POST /api/v1/auth/code
func (h *AuthHandler) CodeLogin(c *fiber.Ctx) error {
	var req models.CodeLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Code is required",
		})
	}

	var user models.User
	// CHANGED: Removed the role = 'student' check to allow admins too
	err := h.DB.DB.QueryRow(`
        SELECT id, role, is_active 
        FROM users 
        WHERE login_code = $1
    `, req.Code).Scan(&user.ID, &user.Role, &user.IsActive)

	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: "Invalid code or inactive user",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Database error",
		})
	}

	if !user.IsActive {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: "Invalid code or inactive user",
		})
	}

	token, err := utils.GenerateJWT(user.ID, user.Role, h.JWTSecret, time.Duration(h.JWTExpiry)*time.Second)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to generate token",
		})
	}

	return c.JSON(models.AuthResponse{
		Token:  token,
		Role:   user.Role,
		UserID: user.ID,
	})
}
