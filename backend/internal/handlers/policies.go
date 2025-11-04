package handlers

import (
	"vote/internal/database"
	"vote/internal/models"

	"github.com/gofiber/fiber/v2"
)

type PolicyHandler struct {
	DB *database.Database
}

func NewPolicyHandler(db *database.Database) *PolicyHandler {
	return &PolicyHandler{DB: db}
}

// GET /api/v1/policies
func (h *PolicyHandler) GetPolicies(c *fiber.Ctx) error {
	// CHANGED: Now includes rejected policies
	rows, err := h.DB.DB.Query(`
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
		WHERE p.status IN ('approved', 'uncertain', 'rejected')
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`)
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

// POST /api/v1/policies
func (h *PolicyHandler) CreatePolicy(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req models.CreatePolicyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validation
	if len(req.Title) < 10 || len(req.Title) > 200 {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Title must be between 10 and 200 characters",
		})
	}

	if len(req.Description) < 50 || len(req.Description) > 2000 {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Description must be between 50 and 2000 characters",
		})
	}

	var policyID string
	err := h.DB.DB.QueryRow(`
		INSERT INTO policies (title, description, submitted_by, status)
		VALUES ($1, $2, $3, 'pending')
		RETURNING id
	`, req.Title, req.Description, userID).Scan(&policyID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to create policy",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.MessageResponse{
		ID:      policyID,
		Status:  "pending",
		Message: "Submitted for review.",
	})
}
