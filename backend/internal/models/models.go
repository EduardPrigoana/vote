package models

import (
	"encoding/json" // ADD THIS LINE
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	LoginCode *string   `json:"login_code,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type Policy struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	AdminComment *string   `json:"admin_comment,omitempty"`
	SubmittedBy  string    `json:"submitted_by"`
	CreatedAt    time.Time `json:"created_at"`
	Upvotes      int       `json:"upvotes,omitempty"`
	Downvotes    int       `json:"downvotes,omitempty"`
}

type Vote struct {
	ID        string    `json:"id"`
	PolicyID  string    `json:"policy_id"`
	UserID    string    `json:"user_id"`
	VoteType  string    `json:"vote_type"`
	CreatedAt time.Time `json:"created_at"`
}

// Request/Response DTOs
type CodeLoginRequest struct {
	Code string `json:"code"`
}

type AuthResponse struct {
	Token  string `json:"token"`
	Role   string `json:"role"`
	UserID string `json:"user_id"`
}

type CreatePolicyRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type VoteRequest struct {
	PolicyID string `json:"policy_id"`
	VoteType string `json:"vote_type"`
}

type UpdateStatusRequest struct {
	Status  string  `json:"status"`
	Comment *string `json:"comment,omitempty"`
}

type CreateUserRequest struct {
	LoginCode string `json:"login_code"`
	IsActive  bool   `json:"is_active"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
	Status  string `json:"status,omitempty"`
}

// Add to existing models

type Category struct {
	ID        string    `json:"id"`
	NameEn    string    `json:"name_en"`
	NameRo    string    `json:"name_ro"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID          string    `json:"id"`
	PolicyID    string    `json:"policy_id"`
	UserID      string    `json:"user_id"`
	CommentText string    `json:"comment_text"`
	IsFlagged   bool      `json:"is_flagged"`
	CreatedAt   time.Time `json:"created_at"`
}

type AuditLogEntry struct {
	ID         string          `json:"id"`
	UserID     *string         `json:"user_id"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   *string         `json:"entity_id"`
	Details    json.RawMessage `json:"details"`
	CreatedAt  time.Time       `json:"created_at"`
}

// Update Policy struct to include new fields
type PolicyExtended struct {
	Policy
	CategoryID          *string `json:"category_id,omitempty"`
	CategoryName        *string `json:"category_name,omitempty"`
	ImplementationDate  *string `json:"implementation_date,omitempty"`
	EstimatedCompletion *string `json:"estimated_completion,omitempty"`
	ViewCount           int     `json:"view_count"`
	CommentCount        int     `json:"comment_count,omitempty"`
}

// Request DTOs
type CreateCommentRequest struct {
	PolicyID    string `json:"policy_id"`
	CommentText string `json:"comment_text"`
}

type UpdatePolicyExtendedRequest struct {
	Status              string  `json:"status"`
	Comment             *string `json:"comment,omitempty"`
	CategoryID          *string `json:"category_id,omitempty"`
	ImplementationDate  *string `json:"implementation_date,omitempty"`
	EstimatedCompletion *string `json:"estimated_completion,omitempty"`
}

type BulkActionRequest struct {
	PolicyIDs  []string `json:"policy_ids"`
	Action     string   `json:"action"`
	Status     *string  `json:"status,omitempty"`
	CategoryID *string  `json:"category_id,omitempty"`
}

type AnalyticsResponse struct {
	TotalPolicies        int                   `json:"total_policies"`
	TotalVotes           int                   `json:"total_votes"`
	TotalComments        int                   `json:"total_comments"`
	ParticipationRate    float64               `json:"participation_rate"`
	VotingTrends         []TrendData           `json:"voting_trends"`
	TopClassrooms        []ClassroomEngagement `json:"top_classrooms"`
	PolicySuccessRate    float64               `json:"policy_success_rate"`
	CategoryDistribution []CategoryStats       `json:"category_distribution"`
}

type TrendData struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type ClassroomEngagement struct {
	LoginCode       string  `json:"login_code"`
	VoteCount       int     `json:"vote_count"`
	CommentCount    int     `json:"comment_count"`
	PolicyCount     int     `json:"policy_count"`
	EngagementScore float64 `json:"engagement_score"`
}

type CategoryStats struct {
	CategoryName string `json:"category_name"`
	PolicyCount  int    `json:"policy_count"`
	VoteCount    int    `json:"vote_count"`
}
