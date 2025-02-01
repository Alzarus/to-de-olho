package controllers

import (
	"net/http"
	"strconv"
	"to-de-olho-api/models"
	"to-de-olho-api/repositories"

	"github.com/gin-gonic/gin"
)

func GetTravelExpenses(c *gin.Context) {
	travelExpenses, err := repositories.GetAllTravelExpenses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching travel expenses"})
		return
	}
	c.JSON(http.StatusOK, travelExpenses)
}

func CreateTravelExpense(c *gin.Context) {
	var travelExpense models.TravelExpense
	if err := c.ShouldBindJSON(&travelExpense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.CreateTravelExpense(&travelExpense); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating travel expense"})
		return
	}

	c.JSON(http.StatusCreated, travelExpense)
}

func CreateTravelExpenses(c *gin.Context) {
	var travelExpenses []models.TravelExpense
	if err := c.ShouldBindJSON(&travelExpenses); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.CreateTravelExpenses(travelExpenses); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating travel expenses"})
		return
	}

	c.JSON(http.StatusCreated, travelExpenses)
}

func GetLatestTravelExpense(c *gin.Context) {
	travelExpense, err := repositories.GetLatestTravelExpense()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No latest travel expense found"})
		return
	}
	c.JSON(http.StatusOK, travelExpense)
}

func UpdateTravelExpense(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updatedData models.TravelExpense
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.UpdateTravelExpense(uint(id), &updatedData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating travel expense"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Travel expense updated successfully"})
}

func DeleteTravelExpense(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := repositories.DeleteTravelExpense(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting travel expense"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Travel expense deleted successfully"})
}
