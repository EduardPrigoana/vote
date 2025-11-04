package handlers

import (
	"database/sql"
	"vote/internal/database"
	"vote/internal/models"

	"github.com/gofiber/fiber/v2"
)

type VoteHandler struct {
	DB *database.Database
}

func NewVoteHandler(db *database.Database) *VoteHandler {
	return &VoteHandler{DB: db}
}

// POST /api/v1/votes
func (h *VoteHandler) CreateVote(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	role := c.Locals("role").(string)

	if role != "student" {
		return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse{
			Error: "Only students can vote",
		})
	}

	var req models.VoteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if req.VoteType != "upvote" && req.VoteType != "downvote" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Vote type must be 'upvote' or 'downvote'",
		})
	}

	// Get device fingerprint from request
	deviceFingerprint := c.Get("X-Device-Fingerprint", "")
	if deviceFingerprint == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Device fingerprint required",
		})
	}

	// Check if policy exists and is votable
	var status string
	err := h.DB.DB.QueryRow(`
		SELECT status FROM policies WHERE id = $1
	`, req.PolicyID).Scan(&status)

	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "Policy not found",
		})
	}

	if status != "approved" && status != "uncertain" && status != "rejected" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Can only vote on approved, uncertain, or rejected policies",
		})
	}

	// Check if this device has already voted on this policy
	var deviceVoteCount int
	err = h.DB.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM votes 
		WHERE policy_id = $1 AND device_fingerprint = $2
	`, req.PolicyID, deviceFingerprint).Scan(&deviceVoteCount)

	if err != nil && err != sql.ErrNoRows {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Database error",
		})
	}

	if deviceVoteCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(models.ErrorResponse{
			Error: "You have already voted on this policy from this device",
		})
	}

	// Insert vote with device fingerprint (NO 100 LIMIT)
	_, err = h.DB.DB.Exec(`
		INSERT INTO votes (policy_id, user_id, vote_type, device_fingerprint)
		VALUES ($1, $2, $3, $4)
	`, req.PolicyID, userID, req.VoteType, deviceFingerprint)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to record vote",
		})
	}

	return c.JSON(models.MessageResponse{
		Message: "Vote recorded.",
	})
}

// GET /api/v1/votes/status/:policyId
func (h *VoteHandler) GetVoteStatus(c *fiber.Ctx) error {
	policyID := c.Params("policyId")
	deviceFingerprint := c.Get("X-Device-Fingerprint", "")

	var deviceHasVoted bool

	if deviceFingerprint != "" {
		var count int
		h.DB.DB.QueryRow(`
			SELECT COUNT(*) 
			FROM votes 
			WHERE policy_id = $1 AND device_fingerprint = $2
		`, policyID, deviceFingerprint).Scan(&count)
		deviceHasVoted = count > 0
	}

	return c.JSON(fiber.Map{
		"device_has_voted": deviceHasVoted,
	})
}
