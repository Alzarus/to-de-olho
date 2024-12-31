package controllers

import (
	"net/http"
	"time"
	"to-de-olho-api/models"
	"to-de-olho-api/repositories"

	"github.com/gin-gonic/gin"
)

func GetExecutionStatus(c *gin.Context) {
	today := time.Now().Format("2006-01-02")
	executionStatus, err := repositories.GetExecutionStatusByDate(today)

	if err != nil {

		c.JSON(http.StatusOK, gin.H{
			"status":  "READY",
			"message": "Nenhuma execução encontrada, status padrão READY",
		})
		return
	}

	c.JSON(http.StatusOK, executionStatus)
}

func LogExecution(c *gin.Context) {
	var executionStatus models.ExecutionStatus
	if err := c.ShouldBindJSON(&executionStatus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if executionStatus.ExecutedAt == "" {
		executionStatus.ExecutedAt = time.Now().Format("2006-01-02")
	}

	err := repositories.UpdateOrCreateExecutionStatus(&executionStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error logging execution"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Execution logged successfully"})
}
