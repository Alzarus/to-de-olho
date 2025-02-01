package controllers

import (
	"net/http"
	"strconv"
	"to-de-olho-api/models"
	"to-de-olho-api/repositories"

	"github.com/gin-gonic/gin"
)

func CreateContract(c *gin.Context) {
	var contract models.Contract
	if err := c.ShouldBindJSON(&contract); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}
	if err := repositories.CreateContract(&contract); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating contract"})
		return
	}
	c.JSON(http.StatusCreated, contract)
}

func CreateContracts(c *gin.Context) {
	var contracts []models.Contract
	if err := c.ShouldBindJSON(&contracts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}
	if err := repositories.CreateContracts(contracts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating contracts"})
		return
	}
	c.JSON(http.StatusCreated, contracts)
}

func GetContracts(c *gin.Context) {
	contracts, err := repositories.GetAllContracts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching contracts"})
		return
	}
	c.JSON(http.StatusOK, contracts)
}

func GetContractByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	contract, err := repositories.GetContractByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contract not found"})
		return
	}
	c.JSON(http.StatusOK, contract)
}

func GetLatestContract(c *gin.Context) {
	contract, err := repositories.GetLatestContract()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No latest contract found"})
		return
	}
	c.JSON(http.StatusOK, contract)
}

func UpdateContract(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updatedData models.Contract
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if err := repositories.UpdateContract(uint(id), &updatedData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating contract"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Contract updated successfully"})
}

func UpdateContracts(c *gin.Context) {
	var updatedContracts []models.Contract
	if err := c.ShouldBindJSON(&updatedContracts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	for _, contract := range updatedContracts {
		if err := repositories.UpdateContract(contract.ID, &contract); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating contract with ID " + strconv.Itoa(int(contract.ID))})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Contracts updated successfully"})
}

func DeleteContract(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := repositories.DeleteContract(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting contract"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Contract deleted successfully"})
}
