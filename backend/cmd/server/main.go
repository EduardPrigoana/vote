package main

import (
	"log"
	"vote/internal/config"
	"vote/internal/database"
	"vote/internal/handlers"
	"vote/internal/middleware"
	"vote/internal/utils"

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

	// Load profanity filter
	utils.LoadProfanityList()

	// Initialize audit logger
	auditLogger := utils.NewAuditLogger(db.DB)

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

	// Serve static assets (CSS, JS, images)
	app.Static("/css", "../frontend/css")
	app.Static("/js", "../frontend/js")

	// Serve HTML pages at clean routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("../frontend/login.html")
	})

	app.Get("/login", func(c *fiber.Ctx) error {
		return c.SendFile("../frontend/login.html")
	})

	app.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.SendFile("../frontend/dashboard.html")
	})

	app.Get("/submit", func(c *fiber.Ctx) error {
		return c.SendFile("../frontend/submit.html")
	})

	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.SendFile("../frontend/admin.html")
	})

	app.Get("/superuser", func(c *fiber.Ctx) error {
		return c.SendFile("../frontend/superuser.html")
	})

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret, int64(cfg.JWTExpiry.Seconds()))
	policyHandler := handlers.NewPolicyHandler(db, auditLogger)
	voteHandler := handlers.NewVoteHandler(db)
	adminHandler := handlers.NewAdminHandler(db, auditLogger)
	superuserHandler := handlers.NewSuperuserHandler(db)
	categoryHandler := handlers.NewCategoryHandler(db)
	analyticsHandler := handlers.NewAnalyticsHandler(db)
	exportHandler := handlers.NewExportHandler(db)

	// API routes
	api := app.Group("/api/v1")

	// Public routes
	api.Post("/auth/code", authHandler.CodeLogin)
	api.Get("/categories", categoryHandler.GetCategories)

	// Protected routes (students + admins + superusers)
	protected := api.Group("", middleware.AuthRequired(cfg.JWTSecret))
	protected.Get("/policies", policyHandler.GetPolicies)
	protected.Get("/policies/:id", policyHandler.GetPolicy)
	protected.Post("/policies", policyHandler.CreatePolicy)
	protected.Post("/votes", voteHandler.CreateVote)
	protected.Get("/votes/status/:policyId", voteHandler.GetVoteStatus)

	// Admin routes (admins + superusers)
	admin := api.Group("/admin", middleware.AuthRequired(cfg.JWTSecret), middleware.AdminRequired())
	admin.Get("/policies", adminHandler.GetAllPolicies)
	admin.Post("/policies/:id/status", adminHandler.UpdatePolicyStatus)
	admin.Post("/policies/:id/comment", adminHandler.AddComment)
	admin.Delete("/policies/:id", adminHandler.DeletePolicy)
	admin.Post("/policies/bulk", adminHandler.BulkAction)
	admin.Post("/users", adminHandler.CreateUser)
	admin.Get("/stats", adminHandler.GetStats)
	admin.Get("/analytics", analyticsHandler.GetAnalytics)
	admin.Get("/audit-log", adminHandler.GetAuditLog)

	// Export
	admin.Get("/export/csv", exportHandler.ExportCSV)
	admin.Get("/export/xlsx", exportHandler.ExportExcel)

	// Superuser routes (superusers only)
	superuser := api.Group("/superuser", middleware.AuthRequired(cfg.JWTSecret), middleware.SuperuserRequired())
	superuser.Get("/users", superuserHandler.GetAllUsers)
	superuser.Post("/users", superuserHandler.CreateUser)
	superuser.Put("/users/:id", superuserHandler.UpdateUser)
	superuser.Delete("/users/:id", superuserHandler.DeleteUser)
	superuser.Post("/users/:id/toggle", superuserHandler.ToggleUserStatus)

	// Start server
	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
