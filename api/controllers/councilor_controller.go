package controllers

import (
	"net/http"
	"strconv"
	"to-de-olho-api/models"
	"to-de-olho-api/repositories"

	"github.com/gin-gonic/gin"
)

func GetCouncilors(c *gin.Context) {
	councilors, err := repositories.GetAllCouncilors()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching councilors"})
		return
	}
	c.JSON(http.StatusOK, councilors)
}

func GetCouncilorByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	councilor, err := repositories.GetCouncilorByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Councilor not found"})
		return
	}
	c.JSON(http.StatusOK, councilor)
}

func CreateCouncilor(c *gin.Context) {
	var councilor models.Councilor
	if err := c.ShouldBindJSON(&councilor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.CreateCouncilor(&councilor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating councilor"})
		return
	}

	c.JSON(http.StatusCreated, councilor)
}

func CreateCouncilors(c *gin.Context) {
	var councilors []models.Councilor
	if err := c.ShouldBindJSON(&councilors); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.CreateCouncilors(councilors); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating councilors"})
		return
	}

	c.JSON(http.StatusCreated, councilors)
}

func UpdateCouncilor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updatedData models.Councilor
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.UpdateCouncilor(uint(id), &updatedData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating councilor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Councilor updated successfully"})
}

func DeleteCouncilor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := repositories.DeleteCouncilor(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting councilor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Councilor deleted successfully"})
}
