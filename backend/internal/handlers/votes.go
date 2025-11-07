package handlers

import (
	"database/sql"
	"vote/internal/database"
	"vote/internal/models"
	"vote/internal/services"

	"github.com/gofiber/fiber/v2"
)

type VoteHandler struct {
	DB    *database.Database
	WSHub *services.WebSocketHub
}

func NewVoteHandler(db *database.Database, wsHub *services.WebSocketHub) *VoteHandler {
	return &VoteHandler{
		DB:    db,
		WSHub: wsHub,
	}
}

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
			Error: "Vote type must be upvote or downvote",
		})
	}

	deviceFingerprint := c.Get("X-Device-Fingerprint", "")
	if deviceFingerprint == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Device fingerprint required",
		})
	}

	var status string
	err := h.DB.DB.QueryRow(`SELECT status FROM policies WHERE id = $1`, req.PolicyID).Scan(&status)

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
			Error: "You have already voted on this policy",
		})
	}

	_, err = h.DB.DB.Exec(`
		INSERT INTO votes (policy_id, user_id, vote_type, device_fingerprint)
		VALUES ($1, $2, $3, $4)
	`, req.PolicyID, userID, req.VoteType, deviceFingerprint)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to record vote",
		})
	}

	var upvotes, downvotes int
	h.DB.DB.QueryRow(`
		SELECT 
			COALESCE(SUM(CASE WHEN vote_type = 'upvote' THEN 1 ELSE 0 END), 0) as upvotes,
			COALESCE(SUM(CASE WHEN vote_type = 'downvote' THEN 1 ELSE 0 END), 0) as downvotes
		FROM votes
		WHERE policy_id = $1
	`, req.PolicyID).Scan(&upvotes, &downvotes)

	if h.WSHub != nil {
		h.WSHub.BroadcastVoteUpdate(req.PolicyID, upvotes, downvotes)
	}

	return c.JSON(models.MessageResponse{
		Message: "Vote recorded",
	})
}
