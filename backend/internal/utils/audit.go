package utils

import (
	"database/sql"
	"encoding/json"
	"log"
)

type AuditLogger struct {
	DB *sql.DB
}

func NewAuditLogger(db *sql.DB) *AuditLogger {
	return &AuditLogger{DB: db}
}

func (a *AuditLogger) Log(userID, action, entityType, entityID string, details interface{}) {
	detailsJSON, _ := json.Marshal(details)

	_, err := a.DB.Exec(`
		INSERT INTO audit_log (user_id, action, entity_type, entity_id, details)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, action, entityType, entityID, detailsJSON)

	if err != nil {
		log.Printf("Failed to write audit log: %v", err)
	}
}
