package models

import (
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
