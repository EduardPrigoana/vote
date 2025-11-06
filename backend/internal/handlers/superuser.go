package handlers

import (
	"database/sql"
	"vote/internal/database"
	"vote/internal/models"

	"github.com/gofiber/fiber/v2"
)

type SuperuserHandler struct {
	DB *database.Database
}

func NewSuperuserHandler(db *database.Database) *SuperuserHandler {
	return &SuperuserHandler{DB: db}
}

// GET /api/v1/superuser/users
func (h *SuperuserHandler) GetAllUsers(c *fiber.Ctx) error {
	rows, err := h.DB.DB.Query(`
		SELECT id, role, login_code, is_active, created_at
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to fetch users",
		})
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var u models.User
		var loginCode sql.NullString
		err := rows.Scan(&u.ID, &u.Role, &loginCode, &u.IsActive, &u.CreatedAt)
		if err != nil {
			continue
		}
		if loginCode.Valid {
			u.LoginCode = &loginCode.String
		}
		users = append(users, u)
	}

	return c.JSON(users)
}

// POST /api/v1/superuser/users
func (h *SuperuserHandler) CreateUser(c *fiber.Ctx) error {
	var req struct {
		Role      string `json:"role"`
		LoginCode string `json:"login_code"`
		IsActive  bool   `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate role
	validRoles := map[string]bool{
		"student":   true,
		"admin":     true,
		"superuser": true,
	}

	if !validRoles[req.Role] {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid role. Must be 'student', 'admin', or 'superuser'",
		})
	}

	if req.LoginCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Login code is required",
		})
	}

	var userID string
	err := h.DB.DB.QueryRow(`
		INSERT INTO users (role, login_code, is_active)
		VALUES ($1, $2, $3)
		RETURNING id
	`, req.Role, req.LoginCode, req.IsActive).Scan(&userID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to create user (code may already exist)",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.MessageResponse{
		ID:      userID,
		Message: "User created successfully",
	})
}

// PUT /api/v1/superuser/users/:id
func (h *SuperuserHandler) UpdateUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	var req struct {
		Role      string `json:"role"`
		LoginCode string `json:"login_code"`
		IsActive  bool   `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate role
	validRoles := map[string]bool{
		"student":   true,
		"admin":     true,
		"superuser": true,
	}

	if !validRoles[req.Role] {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid role",
		})
	}

	result, err := h.DB.DB.Exec(`
		UPDATE users 
		SET role = $1, login_code = $2, is_active = $3
		WHERE id = $4
	`, req.Role, req.LoginCode, req.IsActive, userID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to update user",
		})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "User not found",
		})
	}

	return c.JSON(models.MessageResponse{
		Message: "User updated successfully",
	})
}

// DELETE /api/v1/superuser/users/:id
func (h *SuperuserHandler) DeleteUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	// Prevent deleting yourself
	currentUserID := c.Locals("user_id").(string)
	if userID == currentUserID {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Cannot delete your own account",
		})
	}

	result, err := h.DB.DB.Exec(`DELETE FROM users WHERE id = $1`, userID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to delete user",
		})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "User not found",
		})
	}

	return c.JSON(models.MessageResponse{
		Message: "User deleted successfully",
	})
}

// POST /api/v1/superuser/users/:id/toggle
func (h *SuperuserHandler) ToggleUserStatus(c *fiber.Ctx) error {
	userID := c.Params("id")

	result, err := h.DB.DB.Exec(`
		UPDATE users 
		SET is_active = NOT is_active
		WHERE id = $1
	`, userID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to toggle user status",
		})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "User not found",
		})
	}

	return c.JSON(models.MessageResponse{
		Message: "User status toggled successfully",
	})
}
