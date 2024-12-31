package controllers

import (
	"net/http"
	"strconv"
	"to-de-olho-api/models"
	"to-de-olho-api/repositories"

	"github.com/gin-gonic/gin"
)

func GetPropositionProductivities(c *gin.Context) {
	productivities, err := repositories.GetAllPropositionProductivities()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching proposition productivities"})
		return
	}
	c.JSON(http.StatusOK, productivities)
}

func GetPropositionProductivityByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	productivity, err := repositories.GetPropositionProductivityByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proposition productivity not found"})
		return
	}
	c.JSON(http.StatusOK, productivity)
}

func CreatePropositionProductivity(c *gin.Context) {
	var productivity models.PropositionProductivity
	if err := c.ShouldBindJSON(&productivity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.CreatePropositionProductivity(&productivity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating proposition productivity"})
		return
	}

	c.JSON(http.StatusCreated, productivity)
}

func CreatePropositionProductivities(c *gin.Context) {
	var productivities []models.PropositionProductivity
	if err := c.ShouldBindJSON(&productivities); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.CreatePropositionProductivities(productivities); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating proposition productivities"})
		return
	}

	c.JSON(http.StatusCreated, productivities)
}

func UpdatePropositionProductivity(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updatedData models.PropositionProductivity
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.UpdatePropositionProductivity(uint(id), &updatedData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating proposition productivity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proposition productivity updated successfully"})
}

func DeletePropositionProductivity(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := repositories.DeletePropositionProductivity(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting proposition productivity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proposition productivity deleted successfully"})
}
