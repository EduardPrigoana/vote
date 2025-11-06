package handlers

import (
	"vote/internal/database"
	"vote/internal/models"

	"github.com/gofiber/fiber/v2"
)

type CategoryHandler struct {
	DB *database.Database
}

func NewCategoryHandler(db *database.Database) *CategoryHandler {
	return &CategoryHandler{DB: db}
}

// GET /api/v1/categories
func (h *CategoryHandler) GetCategories(c *fiber.Ctx) error {
	lang := c.Query("lang", "en")

	rows, err := h.DB.DB.Query(`
		SELECT id, name_en, name_ro, slug, created_at
		FROM categories
		ORDER BY name_en ASC
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to fetch categories",
		})
	}
	defer rows.Close()

	categories := []map[string]interface{}{}
	for rows.Next() {
		var cat models.Category
		err := rows.Scan(&cat.ID, &cat.NameEn, &cat.NameRo, &cat.Slug, &cat.CreatedAt)
		if err != nil {
			continue
		}

		name := cat.NameEn
		if lang == "ro" {
			name = cat.NameRo
		}

		categories = append(categories, map[string]interface{}{
			"id":   cat.ID,
			"name": name,
			"slug": cat.Slug,
		})
	}

	return c.JSON(categories)
}
