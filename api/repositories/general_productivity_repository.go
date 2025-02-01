package repositories

import (
	"to-de-olho-api/configs"
	"to-de-olho-api/models"
)

func CreateGeneralProductivity(productivity *models.GeneralProductivity) error {
	return configs.DB.Create(productivity).Error
}

func CreateGeneralProductivities(productivities []models.GeneralProductivity) error {
	return configs.DB.Create(&productivities).Error
}

func GetAllGeneralProductivities() ([]models.GeneralProductivity, error) {
	var productivities []models.GeneralProductivity
	err := configs.DB.Find(&productivities).Error
	return productivities, err
}

func GetGeneralProductivityByID(id uint) (models.GeneralProductivity, error) {
	var productivity models.GeneralProductivity
	err := configs.DB.First(&productivity, id).Error
	return productivity, err
}

func GetLatestGeneralProductivity() (models.GeneralProductivity, error) {
	var productivity models.GeneralProductivity
	err := configs.DB.Order("ano DESC").First(&productivity).Error
	return productivity, err
}

func UpdateGeneralProductivity(id uint, updatedData *models.GeneralProductivity) error {
	return configs.DB.Model(&models.GeneralProductivity{}).Where("id = ?", id).Updates(updatedData).Error
}

func DeleteGeneralProductivity(id uint) error {
	return configs.DB.Delete(&models.GeneralProductivity{}, id).Error
}
