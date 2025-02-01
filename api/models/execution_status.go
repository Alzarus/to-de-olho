package models

import (
	"time"

	"gorm.io/gorm"
)

type ExecutionStatus struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Status     string `json:"status"`      // READY ou RUNNING
	ExecutedAt string `json:"executed_at"` // Data/hora da execução
}

func (e *ExecutionStatus) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ExecutedAt == "" {
		e.ExecutedAt = time.Now().Format("2006-01-02")
	}
	return
}
