package domain

import "time"

type TransactionStatus string

const (
	StatusPending    TransactionStatus = "PENDING"
	StatusExecuting  TransactionStatus = "EXECUTING"
	StatusCommitted  TransactionStatus = "COMMITTED"
	StatusFailed     TransactionStatus = "FAILED"
	StatusRolledBack TransactionStatus = "ROLLED_BACK"
)

type StepStatus string

const (
	StepPending   StepStatus = "PENDING"
	StepCompleted StepStatus = "COMPLETED"
	StepFailed    StepStatus = "FAILED"
)

type Transaction struct {
	ID           string            `json:"id"`
	Status       TransactionStatus `json:"status"`
	Type         string            `json:"type"` // e.g., "RENAME"
	Description  string            `json:"description"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	MetadataJSON string            `json:"metadata_json"`
	Steps        []TransactionStep `json:"steps"`
}

type TransactionStep struct {
	TransactionID string     `json:"transaction_id"`
	StepIndex     int        `json:"step_index"`
	TrackID       uint64     `json:"track_id"`
	Operation     string     `json:"operation"` // "MOVE", "UPDATE_DB"
	SourcePath    string     `json:"source_path"`
	TargetPath    string     `json:"target_path"`
	Status        StepStatus `json:"status"`
	ErrorMessage  string     `json:"error_message"`
}
