package main

import (
	"log"
	"time"
	"vote/internal/config"
	"vote/internal/database"
	"vote/internal/handlers"
	"vote/internal/middleware"
	"vote/internal/services"
	"vote/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	utils.LoadProfanityList()

	auditLogger := utils.NewAuditLogger(db.DB)
	wsHub := services.NewWebSocketHub()
	cache := services.NewCache()

	go wsHub.Run()

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

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middleware.CORS(cfg))
	app.Use(middleware.DomainRestriction(cfg))
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
	}))

	app.Static("/css", "../frontend/css")
	app.Static("/js", "../frontend/js")

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

	app.Get("/policy/:id", func(c *fiber.Ctx) error {
		return c.SendFile("../frontend/policy.html")
	})

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		wsHub.HandleConnection(c)
	}))

	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret, int64(cfg.JWTExpiry.Seconds()))
	policyHandler := handlers.NewPolicyHandler(db, auditLogger, wsHub, cache)
	voteHandler := handlers.NewVoteHandler(db, wsHub)
	adminHandler := handlers.NewAdminHandler(db, auditLogger)
	superuserHandler := handlers.NewSuperuserHandler(db)
	categoryHandler := handlers.NewCategoryHandler(db)
	analyticsHandler := handlers.NewAnalyticsHandler(db)
	exportHandler := handlers.NewExportHandler(db)

	api := app.Group("/api/v1")

	api.Post("/auth/code", authHandler.CodeLogin)
	api.Get("/categories", categoryHandler.GetCategories)

	protected := api.Group("", middleware.AuthRequired(cfg.JWTSecret))
	protected.Get("/policies", policyHandler.GetPolicies)
	protected.Get("/policies/:id", policyHandler.GetPolicy)
	protected.Post("/policies", policyHandler.CreatePolicy)
	protected.Post("/votes", voteHandler.CreateVote)

	admin := api.Group("/admin", middleware.AuthRequired(cfg.JWTSecret), middleware.AdminRequired())
	admin.Get("/policies", adminHandler.GetAllPolicies)
	admin.Get("/policies/:id", adminHandler.GetPolicyForEdit)
	admin.Put("/policies/:id", adminHandler.UpdatePolicy)
	admin.Post("/policies/:id/status", adminHandler.UpdatePolicyStatus)
	admin.Post("/policies/:id/comment", adminHandler.AddComment)
	admin.Delete("/policies/:id", adminHandler.DeletePolicy)
	admin.Post("/policies/bulk", adminHandler.BulkAction)
	admin.Post("/users", adminHandler.CreateUser)
	admin.Get("/stats", adminHandler.GetStats)
	admin.Get("/analytics", analyticsHandler.GetAnalytics)
	admin.Get("/audit-log", adminHandler.GetAuditLog)
	admin.Get("/export/csv", exportHandler.ExportCSV)
	admin.Get("/export/xlsx", exportHandler.ExportExcel)

	superuser := api.Group("/superuser", middleware.AuthRequired(cfg.JWTSecret), middleware.SuperuserRequired())
	superuser.Get("/users", superuserHandler.GetAllUsers)
	superuser.Post("/users", superuserHandler.CreateUser)
	superuser.Put("/users/:id", superuserHandler.UpdateUser)
	superuser.Delete("/users/:id", superuserHandler.DeleteUser)
	superuser.Post("/users/:id/toggle", superuserHandler.ToggleUserStatus)

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
