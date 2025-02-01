package controllers

import (
	"net/http"
	"strconv"
	"to-de-olho-api/models"
	"to-de-olho-api/repositories"

	"github.com/gin-gonic/gin"
)

func CreateGeneralProductivity(c *gin.Context) {
	var productivity models.GeneralProductivity
	if err := c.ShouldBindJSON(&productivity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}
	if err := repositories.CreateGeneralProductivity(&productivity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating general productivity"})
		return
	}
	c.JSON(http.StatusCreated, productivity)
}

func CreateGeneralProductivities(c *gin.Context) {
	var productivities []models.GeneralProductivity
	if err := c.ShouldBindJSON(&productivities); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}
	if err := repositories.CreateGeneralProductivities(productivities); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating general productivities"})
		return
	}
	c.JSON(http.StatusCreated, productivities)
}

func GetGeneralProductivities(c *gin.Context) {
	productivities, err := repositories.GetAllGeneralProductivities()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching general productivities"})
		return
	}
	c.JSON(http.StatusOK, productivities)
}

func GetGeneralProductivityByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	productivity, err := repositories.GetGeneralProductivityByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "General productivity not found"})
		return
	}
	c.JSON(http.StatusOK, productivity)
}

func GetLatestGeneralProductivity(c *gin.Context) {
	productivity, err := repositories.GetLatestGeneralProductivity()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No latest general productivity found"})
		return
	}
	c.JSON(http.StatusOK, productivity)
}

func UpdateGeneralProductivity(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updatedData models.GeneralProductivity
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.UpdateGeneralProductivity(uint(id), &updatedData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating general productivity"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "General productivity updated successfully"})
}

func DeleteGeneralProductivity(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := repositories.DeleteGeneralProductivity(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting general productivity"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "General productivity deleted successfully"})
}
