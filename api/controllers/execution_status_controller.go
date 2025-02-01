package controllers

import (
	"net/http"
	"time"
	"to-de-olho-api/models"
	"to-de-olho-api/repositories"

	"github.com/gin-gonic/gin"
)

// Retorna o último status de execução do dia atual
func GetExecutionStatus(c *gin.Context) {
	today := time.Now().Format("2006-01-02")
	executionStatus, err := repositories.GetExecutionStatusByDate(today)

	// Retorna erro 404 caso não tenha um registro válido
	if err != nil || executionStatus.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "No execution record found",
			"message": "Os crawlers ainda não foram executados hoje.",
		})
		return
	}

	c.JSON(http.StatusOK, executionStatus)
}

// Registra ou atualiza o status de execução
func LogExecution(c *gin.Context) {
	var executionStatus models.ExecutionStatus
	if err := c.ShouldBindJSON(&executionStatus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	// Garante que a data seja sempre do formato correto
	if executionStatus.ExecutedAt == "" {
		executionStatus.ExecutedAt = time.Now().Format("2006-01-02")
	}

	// Atualiza ou insere o status de execução
	err := repositories.UpdateOrCreateExecutionStatus(&executionStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error logging execution"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Execution logged successfully"})
}
