package handlers

import (
	"database/sql"
	"vote/internal/database"
	"vote/internal/models"
	"vote/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type CommentHandler struct {
	DB          *database.Database
	AuditLogger *utils.AuditLogger
}

func NewCommentHandler(db *database.Database, auditLogger *utils.AuditLogger) *CommentHandler {
	return &CommentHandler{
		DB:          db,
		AuditLogger: auditLogger,
	}
}

// GET /api/v1/comments/:policyId
func (h *CommentHandler) GetComments(c *fiber.Ctx) error {
	policyID := c.Params("policyId")

	rows, err := h.DB.DB.Query(`
		SELECT id, policy_id, user_id, comment_text, is_flagged, created_at
		FROM comments
		WHERE policy_id = $1 AND is_flagged = false
		ORDER BY created_at ASC
	`, policyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to fetch comments",
		})
	}
	defer rows.Close()

	comments := []models.Comment{}
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.PolicyID,
			&comment.UserID,
			&comment.CommentText,
			&comment.IsFlagged,
			&comment.CreatedAt,
		)
		if err != nil {
			continue
		}
		comments = append(comments, comment)
	}

	return c.JSON(comments)
}

// POST /api/v1/comments
func (h *CommentHandler) CreateComment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req models.CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate comment length
	if len(req.CommentText) < 1 || len(req.CommentText) > 1000 {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Comment must be between 1 and 1000 characters",
		})
	}

	// Check for profanity
	if utils.ContainsProfanity(req.CommentText) {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Comment contains inappropriate language",
		})
	}

	// Check if policy exists
	var exists bool
	err := h.DB.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM policies WHERE id = $1)`, req.PolicyID).Scan(&exists)
	if err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "Policy not found",
		})
	}

	var commentID string
	err = h.DB.DB.QueryRow(`
		INSERT INTO comments (policy_id, user_id, comment_text)
		VALUES ($1, $2, $3)
		RETURNING id
	`, req.PolicyID, userID, req.CommentText).Scan(&commentID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to create comment",
		})
	}

	// Audit log
	h.AuditLogger.Log(userID, "create_comment", "comment", commentID, map[string]interface{}{
		"policy_id": req.PolicyID,
	})

	return c.Status(fiber.StatusCreated).JSON(models.MessageResponse{
		ID:      commentID,
		Message: "Comment posted successfully",
	})
}

// DELETE /api/v1/comments/:id
func (h *CommentHandler) DeleteComment(c *fiber.Ctx) error {
	commentID := c.Params("id")
	userID := c.Locals("user_id").(string)
	role := c.Locals("role").(string)

	// Check ownership or admin/superuser role
	var ownerID string
	err := h.DB.DB.QueryRow(`SELECT user_id FROM comments WHERE id = $1`, commentID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "Comment not found",
		})
	}

	if ownerID != userID && role != "admin" && role != "superuser" {
		return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse{
			Error: "You can only delete your own comments",
		})
	}

	_, err = h.DB.DB.Exec(`DELETE FROM comments WHERE id = $1`, commentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to delete comment",
		})
	}

	// Audit log
	h.AuditLogger.Log(userID, "delete_comment", "comment", commentID, nil)

	return c.JSON(models.MessageResponse{
		Message: "Comment deleted successfully",
	})
}
