package handlers

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"
	"vote/internal/database"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

type ExportHandler struct {
	DB *database.Database
}

func NewExportHandler(db *database.Database) *ExportHandler {
	return &ExportHandler{DB: db}
}

// GET /api/v1/admin/export/csv
func (h *ExportHandler) ExportCSV(c *fiber.Ctx) error {
	policyIDs := c.Query("ids", "")

	query := `
		SELECT 
			p.id, p.title, p.description, p.status, p.created_at,
			COALESCE(c.name_en, 'N/A') as category,
			COALESCE(SUM(CASE WHEN v.vote_type = 'upvote' THEN 1 ELSE 0 END), 0) as upvotes,
			COALESCE(SUM(CASE WHEN v.vote_type = 'downvote' THEN 1 ELSE 0 END), 0) as downvotes
		FROM policies p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN votes v ON p.id = v.policy_id
	`

	args := []interface{}{}
	if policyIDs != "" {
		ids := strings.Split(policyIDs, ",")
		placeholders := make([]string, len(ids))
		for i, id := range ids {
			args = append(args, id)
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		query += fmt.Sprintf(` WHERE p.id IN (%s)`, strings.Join(placeholders, ","))
	}

	query += ` GROUP BY p.id, c.name_en ORDER BY p.created_at DESC`

	rows, err := h.DB.DB.Query(query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch data",
		})
	}
	defer rows.Close()

	// Create CSV in memory
	var csvData strings.Builder
	writer := csv.NewWriter(&csvData)

	// Header
	writer.Write([]string{"ID", "Title", "Description", "Category", "Status", "Upvotes", "Downvotes", "Submitted"})

	for rows.Next() {
		var id, title, description, category, status string
		var upvotes, downvotes int
		var createdAt time.Time

		rows.Scan(&id, &title, &description, &status, &createdAt, &category, &upvotes, &downvotes)

		writer.Write([]string{
			id,
			title,
			description,
			category,
			status,
			strconv.Itoa(upvotes),
			strconv.Itoa(downvotes),
			createdAt.Format("2006-01-02 15:04"),
		})
	}

	writer.Flush()

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=policies_%s.csv", time.Now().Format("2006-01-02")))

	return c.SendString(csvData.String())
}

// GET /api/v1/admin/export/xlsx
func (h *ExportHandler) ExportExcel(c *fiber.Ctx) error {
	policyIDs := c.Query("ids", "")

	query := `
		SELECT 
			p.id, p.title, p.description, p.status, p.created_at,
			COALESCE(c.name_en, 'N/A') as category,
			COALESCE(SUM(CASE WHEN v.vote_type = 'upvote' THEN 1 ELSE 0 END), 0) as upvotes,
			COALESCE(SUM(CASE WHEN v.vote_type = 'downvote' THEN 1 ELSE 0 END), 0) as downvotes,
			p.admin_comment
		FROM policies p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN votes v ON p.id = v.policy_id
	`

	args := []interface{}{}
	if policyIDs != "" {
		ids := strings.Split(policyIDs, ",")
		placeholders := make([]string, len(ids))
		for i, id := range ids {
			args = append(args, id)
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		query += fmt.Sprintf(` WHERE p.id IN (%s)`, strings.Join(placeholders, ","))
	}

	query += ` GROUP BY p.id, c.name_en ORDER BY p.created_at DESC`

	rows, err := h.DB.DB.Query(query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch data",
		})
	}
	defer rows.Close()

	// Create Excel file
	f := excelize.NewFile()
	sheet := "Policies"
	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// Style header
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#E5E7EB"}},
	})

	// Header
	headers := []string{"ID", "Title", "Description", "Category", "Status", "Upvotes", "Downvotes", "Admin Comment", "Submitted"}
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// Set column widths
	f.SetColWidth(sheet, "A", "A", 12)
	f.SetColWidth(sheet, "B", "B", 30)
	f.SetColWidth(sheet, "C", "C", 50)
	f.SetColWidth(sheet, "D", "D", 20)
	f.SetColWidth(sheet, "E", "E", 15)
	f.SetColWidth(sheet, "F", "G", 10)
	f.SetColWidth(sheet, "H", "H", 30)
	f.SetColWidth(sheet, "I", "I", 20)

	// Data
	rowNum := 2
	for rows.Next() {
		var id, title, description, category, status string
		var adminComment *string
		var upvotes, downvotes int
		var createdAt time.Time

		rows.Scan(&id, &title, &description, &status, &createdAt, &category, &upvotes, &downvotes, &adminComment)

		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowNum), id)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowNum), title)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowNum), description)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowNum), category)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", rowNum), status)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", rowNum), upvotes)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", rowNum), downvotes)
		if adminComment != nil {
			f.SetCellValue(sheet, fmt.Sprintf("H%d", rowNum), *adminComment)
		}
		f.SetCellValue(sheet, fmt.Sprintf("I%d", rowNum), createdAt.Format("2006-01-02 15:04"))

		rowNum++
	}

	// Save to buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate file",
		})
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=policies_%s.xlsx", time.Now().Format("2006-01-02")))

	return c.Send(buffer.Bytes())
}
