package main

import (
	"log"
	"vote/internal/config"
	"vote/internal/database"
	"vote/internal/handlers"
	"vote/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middleware.CORS(cfg.AllowedOrigins))

	// Serve static files (frontend)
	app.Static("/", "../frontend")

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret, int64(cfg.JWTExpiry.Seconds()))
	policyHandler := handlers.NewPolicyHandler(db)
	voteHandler := handlers.NewVoteHandler(db)
	adminHandler := handlers.NewAdminHandler(db)

	// API routes
	api := app.Group("/api/v1")

	// Public routes
	api.Post("/auth/code", authHandler.CodeLogin)

	// Protected routes (students + admins)
	protected := api.Group("", middleware.AuthRequired(cfg.JWTSecret))
	protected.Get("/policies", policyHandler.GetPolicies)
	protected.Post("/policies", policyHandler.CreatePolicy)
	protected.Post("/votes", voteHandler.CreateVote)
	protected.Get("/votes/status/:policyId", voteHandler.GetVoteStatus) // ADD THIS LINE

	// Admin routes
	admin := api.Group("/admin", middleware.AuthRequired(cfg.JWTSecret), middleware.AdminRequired())
	admin.Get("/policies", adminHandler.GetAllPolicies)
	admin.Post("/policies/:id/status", adminHandler.UpdatePolicyStatus)
	admin.Post("/policies/:id/comment", adminHandler.AddComment)
	admin.Post("/users", adminHandler.CreateUser)
	admin.Get("/stats", adminHandler.GetStats)

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
