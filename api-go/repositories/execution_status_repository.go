package repositories

import (
	"errors"
	"time"
	"to-de-olho-api/configs"
	"to-de-olho-api/models"

	"gorm.io/gorm"
)

func UpdateOrCreateExecutionStatus(executionStatus *models.ExecutionStatus) error {
	today := time.Now().Format("2006-01-02")
	var existingStatus models.ExecutionStatus

	if executionStatus.ExecutedAt == "" {
		executionStatus.ExecutedAt = today
	}

	err := configs.DB.Where("DATE(executed_at) = ?", today).First(&existingStatus).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return configs.DB.Create(executionStatus).Error
	} else if err != nil {
		return err
	}

	existingStatus.Status = executionStatus.Status
	existingStatus.ExecutedAt = executionStatus.ExecutedAt
	return configs.DB.Save(&existingStatus).Error
}

func GetLastExecutionStatus(status string) (models.ExecutionStatus, error) {
	var executionStatus models.ExecutionStatus
	err := configs.DB.Where("status = ?", status).Order("executed_at DESC").First(&executionStatus).Error
	return executionStatus, err
}

func GetExecutionStatusByDate(date string) (models.ExecutionStatus, error) {
	var executionStatus models.ExecutionStatus

	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	err := configs.DB.Where("DATE(executed_at) = ?", date).First(&executionStatus).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newExecution := models.ExecutionStatus{
			Status: "READY",
		}
		if err := configs.DB.Create(&newExecution).Error; err != nil {
			return executionStatus, err
		}
		return newExecution, nil
	}

	if err != nil {
		return executionStatus, errors.New("execution status not found or invalid date")
	}
	return executionStatus, nil
}
