package repositories

import (
	"to-de-olho-api/configs"
	"to-de-olho-api/models"
)

func GetAllCouncilors() ([]models.Councilor, error) {
	var councilors []models.Councilor
	result := configs.DB.Find(&councilors)
	return councilors, result.Error
}

func GetCouncilorByID(id uint) (models.Councilor, error) {
	var councilor models.Councilor
	result := configs.DB.First(&councilor, id)
	return councilor, result.Error
}

func CreateCouncilor(councilor *models.Councilor) error {
	result := configs.DB.Create(councilor)
	return result.Error
}

func CreateCouncilors(councilors []models.Councilor) error {
	result := configs.DB.Create(&councilors)
	return result.Error
}

func UpdateCouncilor(id uint, updatedData *models.Councilor) error {
	result := configs.DB.Model(&models.Councilor{}).Where("id = ?", id).Updates(updatedData)
	return result.Error
}

func DeleteCouncilor(id uint) error {
	result := configs.DB.Delete(&models.Councilor{}, id)
	return result.Error
}
