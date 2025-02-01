package repositories

import (
	"to-de-olho-api/configs"
	"to-de-olho-api/models"
)

func GetAllTravelExpenses() ([]models.TravelExpense, error) {
	var travelExpenses []models.TravelExpense
	result := configs.DB.Find(&travelExpenses)
	return travelExpenses, result.Error
}

func CreateTravelExpense(travelExpense *models.TravelExpense) error {
	result := configs.DB.Create(travelExpense)
	return result.Error
}

func CreateTravelExpenses(travelExpenses []models.TravelExpense) error {
	return configs.DB.Create(&travelExpenses).Error
}

func GetLatestTravelExpense() (models.TravelExpense, error) {
	var travelExpense models.TravelExpense
	result := configs.DB.Order("data DESC").First(&travelExpense)
	return travelExpense, result.Error
}

func UpdateTravelExpense(id uint, updatedData *models.TravelExpense) error {
	result := configs.DB.Model(&models.TravelExpense{}).Where("id = ?", id).Updates(updatedData)
	return result.Error
}

func DeleteTravelExpense(id uint) error {
	result := configs.DB.Delete(&models.TravelExpense{}, id)
	return result.Error
}
