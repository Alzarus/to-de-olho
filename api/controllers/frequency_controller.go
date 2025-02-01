package controllers

import (
	"net/http"
	"strconv"
	"to-de-olho-api/models"
	"to-de-olho-api/repositories"

	"github.com/gin-gonic/gin"
)

func GetFrequencies(c *gin.Context) {
	frequencies, err := repositories.GetAllFrequencies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching frequencies"})
		return
	}
	c.JSON(http.StatusOK, frequencies)
}

func GetFrequencyByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	frequency, err := repositories.GetFrequencyByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Frequency not found"})
		return
	}
	c.JSON(http.StatusOK, frequency)
}

func GetLatestFrequency(c *gin.Context) {
	frequency, err := repositories.GetLatestFrequency()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No latest frequency found"})
		return
	}
	c.JSON(http.StatusOK, frequency)
}

func CreateFrequency(c *gin.Context) {
	var frequency models.Frequency
	if err := c.ShouldBindJSON(&frequency); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.CreateFrequency(&frequency); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating frequency"})
		return
	}

	c.JSON(http.StatusCreated, frequency)
}

func CreateFrequencies(c *gin.Context) {
	var frequencies []models.Frequency
	if err := c.ShouldBindJSON(&frequencies); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.CreateFrequencies(frequencies); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating frequencies"})
		return
	}

	c.JSON(http.StatusCreated, frequencies)
}

func UpdateFrequency(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updatedData models.Frequency
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.UpdateFrequency(uint(id), &updatedData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating frequency"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Frequency updated successfully"})
}

func DeleteFrequency(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := repositories.DeleteFrequency(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting frequency"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Frequency deleted successfully"})
}
