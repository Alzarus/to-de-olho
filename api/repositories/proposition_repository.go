package repositories

import (
	"to-de-olho-api/configs"
	"to-de-olho-api/models"
)

func GetAllPropositions() ([]models.Proposition, error) {
	var propositions []models.Proposition
	result := configs.DB.Find(&propositions)
	return propositions, result.Error
}

func GetPropositionByID(id uint) (models.Proposition, error) {
	var proposition models.Proposition
	result := configs.DB.First(&proposition, id)
	return proposition, result.Error
}

func GetLatestProposition() (models.Proposition, error) {
	var proposition models.Proposition
	result := configs.DB.Order("data_movimentacao DESC").First(&proposition)
	return proposition, result.Error
}

func CreateProposition(proposition *models.Proposition) error {
	result := configs.DB.Create(proposition)
	return result.Error
}

func CreatePropositions(propositions []models.Proposition) error {
	result := configs.DB.Create(&propositions)
	return result.Error
}

func UpdateProposition(id uint, updatedData *models.Proposition) error {
	result := configs.DB.Model(&models.Proposition{}).Where("id = ?", id).Updates(updatedData)
	return result.Error
}

func DeleteProposition(id uint) error {
	result := configs.DB.Delete(&models.Proposition{}, id)
	return result.Error
}
