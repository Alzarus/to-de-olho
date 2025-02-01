package controllers

import (
	"net/http"
	"strconv"
	"to-de-olho-api/models"
	"to-de-olho-api/repositories"

	"github.com/gin-gonic/gin"
)

func GetPropositions(c *gin.Context) {
	propositions, err := repositories.GetAllPropositions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching propositions"})
		return
	}
	c.JSON(http.StatusOK, propositions)
}

func GetPropositionByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	proposition, err := repositories.GetPropositionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proposition not found"})
		return
	}
	c.JSON(http.StatusOK, proposition)
}

func GetLatestProposition(c *gin.Context) {
	proposition, err := repositories.GetLatestProposition()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No latest proposition found"})
		return
	}
	c.JSON(http.StatusOK, proposition)
}

func CreateProposition(c *gin.Context) {
	var proposition models.Proposition
	if err := c.ShouldBindJSON(&proposition); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.CreateProposition(&proposition); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating proposition"})
		return
	}

	c.JSON(http.StatusCreated, proposition)
}

func CreatePropositions(c *gin.Context) {
	var propositions []models.Proposition
	if err := c.ShouldBindJSON(&propositions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.CreatePropositions(propositions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating propositions"})
		return
	}

	c.JSON(http.StatusCreated, propositions)
}

func UpdateProposition(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updatedData models.Proposition
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.UpdateProposition(uint(id), &updatedData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating proposition"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proposition updated successfully"})
}

func DeleteProposition(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := repositories.DeleteProposition(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting proposition"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proposition deleted successfully"})
}
