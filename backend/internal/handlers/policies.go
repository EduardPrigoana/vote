package handlers

import (
	"database/sql"
	"fmt"
	"vote/internal/database"
	"vote/internal/models"
	"vote/internal/services"
	"vote/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type PolicyHandler struct {
	DB          *database.Database
	AuditLogger *utils.AuditLogger
	WSHub       *services.WebSocketHub
	Cache       *services.Cache
}

func NewPolicyHandler(db *database.Database, auditLogger *utils.AuditLogger, wsHub *services.WebSocketHub, cache *services.Cache) *PolicyHandler {
	return &PolicyHandler{
		DB:          db,
		AuditLogger: auditLogger,
		WSHub:       wsHub,
		Cache:       cache,
	}
}

func (h *PolicyHandler) GetPolicies(c *fiber.Ctx) error {
	search := c.Query("search", "")
	status := c.Query("status", "")
	categoryID := c.Query("category", "")
	sortBy := c.Query("sort", "newest")

	deviceFingerprint := c.Get("X-Device-Fingerprint", "")

	query := `
		SELECT 
			p.id, p.title, p.description, p.status, p.admin_comment,
			p.submitted_by, p.created_at, p.category_id, p.view_count,
			c.name_en as category_name,
			COALESCE(SUM(CASE WHEN v.vote_type = 'upvote' THEN 1 ELSE 0 END), 0) as upvotes,
			COALESCE(SUM(CASE WHEN v.vote_type = 'downvote' THEN 1 ELSE 0 END), 0) as downvotes,
			EXISTS(
				SELECT 1 FROM votes 
				WHERE policy_id = p.id AND device_fingerprint = $1
			) as current_user_vote
		FROM policies p
		LEFT JOIN votes v ON p.id = v.policy_id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.status IN ('approved', 'uncertain', 'rejected', 'in_progress', 'completed', 'on_hold', 'cannot_implement')
	`

	args := []interface{}{deviceFingerprint}
	argIndex := 2

	if search != "" {
		query += fmt.Sprintf(` AND (p.title ILIKE $%d OR p.description ILIKE $%d)`, argIndex, argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if status != "" {
		query += fmt.Sprintf(` AND p.status = $%d`, argIndex)
		args = append(args, status)
		argIndex++
	}

	if categoryID != "" {
		query += fmt.Sprintf(` AND p.category_id = $%d`, argIndex)
		args = append(args, categoryID)
		argIndex++
	}

	query += ` GROUP BY p.id, c.name_en`

	switch sortBy {
	case "oldest":
		query += ` ORDER BY p.created_at ASC`
	case "most_voted":
		query += ` ORDER BY (COALESCE(SUM(CASE WHEN v.vote_type = 'upvote' THEN 1 ELSE 0 END), 0) + COALESCE(SUM(CASE WHEN v.vote_type = 'downvote' THEN 1 ELSE 0 END), 0)) DESC`
	case "trending":
		query += ` ORDER BY (COALESCE(SUM(CASE WHEN v.vote_type = 'upvote' THEN 1 ELSE 0 END), 0) + COALESCE(SUM(CASE WHEN v.vote_type = 'downvote' THEN 1 ELSE 0 END), 0)) DESC, p.created_at DESC`
	default:
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
		var currentUserVote bool

		err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.Status, &p.AdminComment,
			&p.SubmittedBy, &p.CreatedAt, &p.CategoryID, &p.ViewCount,
			&categoryName, &p.Upvotes, &p.Downvotes, &currentUserVote,
		)
		if err != nil {
			continue
		}

		if categoryName.Valid {
			p.CategoryName = &categoryName.String
		}

		policyMap := map[string]interface{}{
			"id":                p.ID,
			"title":             p.Title,
			"description":       p.Description,
			"status":            p.Status,
			"admin_comment":     p.AdminComment,
			"upvotes":           p.Upvotes,
			"downvotes":         p.Downvotes,
			"view_count":        p.ViewCount,
			"created_at":        p.CreatedAt,
			"current_user_vote": currentUserVote,
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

func (h *PolicyHandler) GetPolicy(c *fiber.Ctx) error {
	policyID := c.Params("id")
	deviceFingerprint := c.Get("X-Device-Fingerprint", "")

	h.DB.DB.Exec(`UPDATE policies SET view_count = view_count + 1 WHERE id = $1`, policyID)

	query := `
		SELECT 
			p.id, p.title, p.description, p.status, p.admin_comment, p.submitted_by,
			p.created_at, p.category_id, p.view_count,
			COALESCE(SUM(CASE WHEN v.vote_type = 'upvote' THEN 1 ELSE 0 END), 0) as upvotes,
			COALESCE(SUM(CASE WHEN v.vote_type = 'downvote' THEN 1 ELSE 0 END), 0) as downvotes,
			EXISTS(
				SELECT 1 FROM votes 
				WHERE policy_id = p.id AND device_fingerprint = $2
			) as current_user_vote,
			c.name_en as category_name
		FROM policies p
		LEFT JOIN votes v ON p.id = v.policy_id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1
		GROUP BY p.id, c.name_en
	`

	var p models.PolicyExtended
	var currentUserVote bool
	var categoryName sql.NullString

	err := h.DB.DB.QueryRow(query, policyID, deviceFingerprint).Scan(
		&p.ID, &p.Title, &p.Description, &p.Status, &p.AdminComment, &p.SubmittedBy,
		&p.CreatedAt, &p.CategoryID, &p.ViewCount,
		&p.Upvotes, &p.Downvotes, &currentUserVote, &categoryName,
	)

	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "Policy not found",
		})
	}

	if categoryName.Valid {
		p.CategoryName = &categoryName.String
	}

	response := map[string]interface{}{
		"id":                p.ID,
		"title":             p.Title,
		"description":       p.Description,
		"status":            p.Status,
		"admin_comment":     p.AdminComment,
		"upvotes":           p.Upvotes,
		"downvotes":         p.Downvotes,
		"view_count":        p.ViewCount,
		"created_at":        p.CreatedAt,
		"current_user_vote": currentUserVote,
		"category_name":     p.CategoryName,
		"category_id":       p.CategoryID,
	}

	return c.JSON(response)
}

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

	if h.AuditLogger != nil {
		h.AuditLogger.Log(userID, "create_policy", "policy", policyID, map[string]interface{}{
			"title": req.Title,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.MessageResponse{
		ID:      policyID,
		Status:  "pending",
		Message: "Submitted for review",
	})
}
