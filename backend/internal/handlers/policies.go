package handlers

import (
	"database/sql"
	"fmt"
	"vote/internal/database"
	"vote/internal/models"
	"vote/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type PolicyHandler struct {
	DB          *database.Database
	AuditLogger *utils.AuditLogger
}

func NewPolicyHandler(db *database.Database, auditLogger *utils.AuditLogger) *PolicyHandler {
	return &PolicyHandler{
		DB:          db,
		AuditLogger: auditLogger,
	}
}

// GET /api/v1/policies
func (h *PolicyHandler) GetPolicies(c *fiber.Ctx) error {
	search := c.Query("search", "")
	status := c.Query("status", "")
	categoryID := c.Query("category", "")
	sortBy := c.Query("sort", "newest")
	lang := c.Query("lang", "en")

	query := `
		SELECT 
			p.id, 
			p.title, 
			p.description, 
			p.status,
			p.admin_comment,
			p.submitted_by,
			p.created_at,
			p.category_id,
			p.view_count,
			CASE WHEN $1 = 'ro' THEN c.name_ro ELSE c.name_en END as category_name,
			COALESCE(SUM(CASE WHEN v.vote_type = 'upvote' THEN 1 ELSE 0 END), 0) as upvotes,
			COALESCE(SUM(CASE WHEN v.vote_type = 'downvote' THEN 1 ELSE 0 END), 0) as downvotes
		FROM policies p
		LEFT JOIN votes v ON p.id = v.policy_id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.status IN ('approved', 'uncertain', 'rejected', 'in_progress', 'completed', 'on_hold', 'cannot_implement')
	`

	args := []interface{}{lang}
	argIndex := 2

	// Search filter
	if search != "" {
		query += fmt.Sprintf(` AND (p.title ILIKE $%d OR p.description ILIKE $%d)`, argIndex, argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	// Status filter
	if status != "" {
		query += fmt.Sprintf(` AND p.status = $%d`, argIndex)
		args = append(args, status)
		argIndex++
	}

	// Category filter
	if categoryID != "" {
		query += fmt.Sprintf(` AND p.category_id = $%d`, argIndex)
		args = append(args, categoryID)
		argIndex++
	}

	query += ` GROUP BY p.id, c.name_en, c.name_ro`

	// Sorting
	switch sortBy {
	case "oldest":
		query += ` ORDER BY p.created_at ASC`
	case "most_voted":
		query += ` ORDER BY (COALESCE(SUM(CASE WHEN v.vote_type = 'upvote' THEN 1 ELSE 0 END), 0) + COALESCE(SUM(CASE WHEN v.vote_type = 'downvote' THEN 1 ELSE 0 END), 0)) DESC`
	case "trending":
		query += ` ORDER BY (COALESCE(SUM(CASE WHEN v.vote_type = 'upvote' THEN 1 ELSE 0 END), 0) + COALESCE(SUM(CASE WHEN v.vote_type = 'downvote' THEN 1 ELSE 0 END), 0)) DESC, p.created_at DESC`
	default: // newest
		query += ` ORDER BY p.created_at DESC`
	}

	rows, err := h.DB.DB.Query(query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to fetch policies",
		})
	}
	defer rows.Close()

	policies := []map[string]interface{}{}
	for rows.Next() {
		var p models.PolicyExtended
		var categoryName sql.NullString

		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Description,
			&p.Status,
			&p.AdminComment,
			&p.SubmittedBy,
			&p.CreatedAt,
			&p.CategoryID,
			&p.ViewCount,
			&categoryName,
			&p.Upvotes,
			&p.Downvotes,
		)
		if err != nil {
			continue
		}

		if categoryName.Valid {
			p.CategoryName = &categoryName.String
		}

		policyMap := map[string]interface{}{
			"id":            p.ID,
			"title":         p.Title,
			"description":   p.Description,
			"status":        p.Status,
			"admin_comment": p.AdminComment,
			"upvotes":       p.Upvotes,
			"downvotes":     p.Downvotes,
			"view_count":    p.ViewCount,
			"created_at":    p.CreatedAt,
		}

		if p.CategoryName != nil {
			policyMap["category_name"] = *p.CategoryName
		}
		if p.CategoryID != nil {
			policyMap["category_id"] = *p.CategoryID
		}

		policies = append(policies, policyMap)
	}

	return c.JSON(policies)
}

// GET /api/v1/policies/:id
func (h *PolicyHandler) GetPolicy(c *fiber.Ctx) error {
	policyID := c.Params("id")

	// Increment view count
	h.DB.DB.Exec(`UPDATE policies SET view_count = view_count + 1 WHERE id = $1`, policyID)

	query := `
		SELECT 
			p.id, p.title, p.description, p.status, p.admin_comment, p.submitted_by,
			p.created_at, p.category_id, p.view_count, p.implementation_date, p.estimated_completion,
			COALESCE(SUM(CASE WHEN v.vote_type = 'upvote' THEN 1 ELSE 0 END), 0) as upvotes,
			COALESCE(SUM(CASE WHEN v.vote_type = 'downvote' THEN 1 ELSE 0 END), 0) as downvotes
		FROM policies p
		LEFT JOIN votes v ON p.id = v.policy_id
		WHERE p.id = $1
		GROUP BY p.id
	`

	var p models.PolicyExtended
	err := h.DB.DB.QueryRow(query, policyID).Scan(
		&p.ID, &p.Title, &p.Description, &p.Status, &p.AdminComment, &p.SubmittedBy,
		&p.CreatedAt, &p.CategoryID, &p.ViewCount, &p.ImplementationDate, &p.EstimatedCompletion,
		&p.Upvotes, &p.Downvotes,
	)

	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "Policy not found",
		})
	}

	return c.JSON(p)
}

// POST /api/v1/policies
func (h *PolicyHandler) CreatePolicy(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		CategoryID  *string `json:"category_id"`
	}

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

	// Check for profanity
	if utils.ContainsProfanity(req.Title) || utils.ContainsProfanity(req.Description) {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Content contains inappropriate language",
		})
	}

	var policyID string
	err := h.DB.DB.QueryRow(`
		INSERT INTO policies (title, description, submitted_by, status, category_id)
		VALUES ($1, $2, $3, 'pending', $4)
		RETURNING id
	`, req.Title, req.Description, userID, req.CategoryID).Scan(&policyID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to create policy",
		})
	}

	// Audit log
	h.AuditLogger.Log(userID, "create_policy", "policy", policyID, map[string]interface{}{
		"title": req.Title,
	})

	return c.Status(fiber.StatusCreated).JSON(models.MessageResponse{
		ID:      policyID,
		Status:  "pending",
		Message: "Submitted for review.",
	})
}
