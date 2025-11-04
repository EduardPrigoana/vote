package handlers

import (
	"vote/internal/database"
	"vote/internal/models"

	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	DB *database.Database
}

func NewAdminHandler(db *database.Database) *AdminHandler {
	return &AdminHandler{DB: db}
}

// GET /api/v1/admin/policies
func (h *AdminHandler) GetAllPolicies(c *fiber.Ctx) error {
	status := c.Query("status")

	query := `
		SELECT 
			p.id, 
			p.title, 
			p.description, 
			p.status,
			p.admin_comment,
			p.submitted_by,
			p.created_at,
			COALESCE(SUM(CASE WHEN v.vote_type = 'upvote' THEN 1 ELSE 0 END), 0) as upvotes,
			COALESCE(SUM(CASE WHEN v.vote_type = 'downvote' THEN 1 ELSE 0 END), 0) as downvotes
		FROM policies p
		LEFT JOIN votes v ON p.id = v.policy_id
	`

	args := []interface{}{}
	if status != "" {
		query += " WHERE p.status = $1"
		args = append(args, status)
	}

	query += " GROUP BY p.id ORDER BY p.created_at DESC"

	rows, err := h.DB.DB.Query(query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to fetch policies",
		})
	}
	defer rows.Close()

	policies := []models.Policy{}
	for rows.Next() {
		var p models.Policy
		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Description,
			&p.Status,
			&p.AdminComment,
			&p.SubmittedBy,
			&p.CreatedAt,
			&p.Upvotes,
			&p.Downvotes,
		)
		if err != nil {
			continue
		}
		policies = append(policies, p)
	}

	return c.JSON(policies)
}

// POST /api/v1/admin/policies/:id/status
func (h *AdminHandler) UpdatePolicyStatus(c *fiber.Ctx) error {
	policyID := c.Params("id")

	var req models.UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	validStatuses := map[string]bool{
		"pending":   true,
		"approved":  true,
		"rejected":  true,
		"uncertain": true,
	}

	if !validStatuses[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid status",
		})
	}

	result, err := h.DB.DB.Exec(`
		UPDATE policies 
		SET status = $1, admin_comment = $2
		WHERE id = $3
	`, req.Status, req.Comment, policyID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to update policy",
		})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "Policy not found",
		})
	}

	return c.JSON(models.MessageResponse{
		Message: "Policy updated successfully",
	})
}

// POST /api/v1/admin/policies/:id/comment
func (h *AdminHandler) AddComment(c *fiber.Ctx) error {
	policyID := c.Params("id")

	var req struct {
		Comment string `json:"comment"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	result, err := h.DB.DB.Exec(`
		UPDATE policies 
		SET admin_comment = $1
		WHERE id = $2
	`, req.Comment, policyID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to add comment",
		})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "Policy not found",
		})
	}

	return c.JSON(models.MessageResponse{
		Message: "Comment added successfully",
	})
}

// DELETE /api/v1/admin/policies/:id
func (h *AdminHandler) DeletePolicy(c *fiber.Ctx) error {
	policyID := c.Params("id")

	// Delete policy (CASCADE will handle votes)
	result, err := h.DB.DB.Exec(`
		DELETE FROM policies WHERE id = $1
	`, policyID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to delete policy",
		})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "Policy not found",
		})
	}

	return c.JSON(models.MessageResponse{
		Message: "Policy deleted successfully",
	})
}

// POST /api/v1/admin/users
func (h *AdminHandler) CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
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
		VALUES ('student', $1, $2)
		RETURNING id
	`, req.LoginCode, req.IsActive).Scan(&userID)

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

// GET /api/v1/admin/stats
func (h *AdminHandler) GetStats(c *fiber.Ctx) error {
	var stats struct {
		TotalPolicies   int `json:"total_policies"`
		PendingPolicies int `json:"pending_policies"`
		TotalVotes      int `json:"total_votes"`
		ActiveStudents  int `json:"active_students"`
	}

	h.DB.DB.QueryRow("SELECT COUNT(*) FROM policies").Scan(&stats.TotalPolicies)
	h.DB.DB.QueryRow("SELECT COUNT(*) FROM policies WHERE status = 'pending'").Scan(&stats.PendingPolicies)
	h.DB.DB.QueryRow("SELECT COUNT(*) FROM votes").Scan(&stats.TotalVotes)
	h.DB.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'student' AND is_active = true").Scan(&stats.ActiveStudents)

	return c.JSON(stats)
}
