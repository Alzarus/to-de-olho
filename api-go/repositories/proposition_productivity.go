package repositories

import (
	"to-de-olho-api/configs"
	"to-de-olho-api/models"
)

func CreatePropositionProductivity(productivity *models.PropositionProductivity) error {
	return configs.DB.Create(productivity).Error
}

func CreatePropositionProductivities(productivities []models.PropositionProductivity) error {
	return configs.DB.Create(&productivities).Error
}

func GetAllPropositionProductivities() ([]models.PropositionProductivity, error) {
	var productivities []models.PropositionProductivity
	err := configs.DB.Find(&productivities).Error
	return productivities, err
}

func GetPropositionProductivityByID(id uint) (models.PropositionProductivity, error) {
	var productivity models.PropositionProductivity
	err := configs.DB.First(&productivity, id).Error
	return productivity, err
}

func GetLatestPropositionProductivity() (models.PropositionProductivity, error) {
	var productivity models.PropositionProductivity
	err := configs.DB.Order("ano DESC").First(&productivity).Error
	return productivity, err
}

func UpdatePropositionProductivity(id uint, updatedData *models.PropositionProductivity) error {
	return configs.DB.Model(&models.PropositionProductivity{}).Where("id = ?", id).Updates(updatedData).Error
}

func DeletePropositionProductivity(id uint) error {
	return configs.DB.Delete(&models.PropositionProductivity{}, id).Error
}
