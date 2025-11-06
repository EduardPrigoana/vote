package handlers

import (
	"vote/internal/database"
	"vote/internal/models"

	"github.com/gofiber/fiber/v2"
)

type AnalyticsHandler struct {
	DB *database.Database
}

func NewAnalyticsHandler(db *database.Database) *AnalyticsHandler {
	return &AnalyticsHandler{DB: db}
}

// GET /api/v1/admin/analytics
func (h *AnalyticsHandler) GetAnalytics(c *fiber.Ctx) error {
	var analytics models.AnalyticsResponse

	// Total counts
	h.DB.DB.QueryRow(`SELECT COUNT(*) FROM policies`).Scan(&analytics.TotalPolicies)
	h.DB.DB.QueryRow(`SELECT COUNT(*) FROM votes`).Scan(&analytics.TotalVotes)
	analytics.TotalComments = 0 // Set to 0 since we removed comments

	// Participation rate
	var totalStudents int
	h.DB.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'student' AND is_active = true`).Scan(&totalStudents)

	var activeVoters int
	h.DB.DB.QueryRow(`SELECT COUNT(DISTINCT user_id) FROM votes`).Scan(&activeVoters)

	if totalStudents > 0 {
		analytics.ParticipationRate = float64(activeVoters) / float64(totalStudents) * 100
	}

	// Policy success rate
	var totalNonPending int
	h.DB.DB.QueryRow(`
		SELECT COUNT(*) FROM policies 
		WHERE status IN ('approved', 'in_progress', 'completed', 'rejected', 'cannot_implement')
	`).Scan(&totalNonPending)

	var successfulPolicies int
	h.DB.DB.QueryRow(`
		SELECT COUNT(*) FROM policies 
		WHERE status IN ('approved', 'in_progress', 'completed')
	`).Scan(&successfulPolicies)

	if totalNonPending > 0 {
		analytics.PolicySuccessRate = float64(successfulPolicies) / float64(totalNonPending) * 100
	}

	// Voting trends (last 30 days)
	rows, _ := h.DB.DB.Query(`
		SELECT DATE(created_at) as vote_date, COUNT(*) as vote_count
		FROM votes
		WHERE created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY vote_date DESC
		LIMIT 30
	`)
	defer rows.Close()

	trends := []models.TrendData{}
	for rows.Next() {
		var trend models.TrendData
		rows.Scan(&trend.Date, &trend.Count)
		trends = append(trends, trend)
	}
	analytics.VotingTrends = trends

	// Top classrooms by engagement (removed comment count)
	classroomRows, _ := h.DB.DB.Query(`
		SELECT 
			u.login_code,
			COUNT(DISTINCT v.id) as vote_count,
			0 as comment_count,
			COUNT(DISTINCT p.id) as policy_count,
			(COUNT(DISTINCT v.id) + COUNT(DISTINCT p.id) * 3) as engagement_score
		FROM users u
		LEFT JOIN votes v ON u.id = v.user_id
		LEFT JOIN policies p ON u.id = p.submitted_by
		WHERE u.role = 'student' AND u.is_active = true
		GROUP BY u.id, u.login_code
		HAVING COUNT(DISTINCT v.id) > 0 OR COUNT(DISTINCT p.id) > 0
		ORDER BY engagement_score DESC
		LIMIT 10
	`)
	defer classroomRows.Close()

	topClassrooms := []models.ClassroomEngagement{}
	for classroomRows.Next() {
		var classroom models.ClassroomEngagement
		classroomRows.Scan(
			&classroom.LoginCode,
			&classroom.VoteCount,
			&classroom.CommentCount,
			&classroom.PolicyCount,
			&classroom.EngagementScore,
		)
		topClassrooms = append(topClassrooms, classroom)
	}
	analytics.TopClassrooms = topClassrooms

	// Category distribution
	catRows, _ := h.DB.DB.Query(`
		SELECT 
			COALESCE(c.name_en, 'Uncategorized') as category_name,
			COUNT(DISTINCT p.id) as policy_count,
			COUNT(DISTINCT v.id) as vote_count
		FROM policies p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN votes v ON p.id = v.policy_id
		GROUP BY c.name_en
		ORDER BY policy_count DESC
	`)
	defer catRows.Close()

	categoryStats := []models.CategoryStats{}
	for catRows.Next() {
		var stat models.CategoryStats
		catRows.Scan(&stat.CategoryName, &stat.PolicyCount, &stat.VoteCount)
		categoryStats = append(categoryStats, stat)
	}
	analytics.CategoryDistribution = categoryStats

	return c.JSON(analytics)
}
