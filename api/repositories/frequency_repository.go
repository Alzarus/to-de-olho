package repositories

import (
	"to-de-olho-api/configs"
	"to-de-olho-api/models"
)

func GetAllFrequencies() ([]models.Frequency, error) {
	var frequencies []models.Frequency
	result := configs.DB.Find(&frequencies)
	return frequencies, result.Error
}

func GetFrequencyByID(id uint) (models.Frequency, error) {
	var frequency models.Frequency
	result := configs.DB.First(&frequency, id)
	return frequency, result.Error
}

func GetLatestFrequency() (models.Frequency, error) {
	var frequency models.Frequency
	result := configs.DB.Order("ano_sessao DESC, numero_sessao DESC").First(&frequency)
	return frequency, result.Error
}

func CreateFrequency(frequency *models.Frequency) error {
	result := configs.DB.Create(frequency)
	return result.Error
}

func CreateFrequencies(frequencies []models.Frequency) error {
	result := configs.DB.Create(&frequencies)
	return result.Error
}

func UpdateFrequency(id uint, updatedData *models.Frequency) error {
	result := configs.DB.Model(&models.Frequency{}).Where("id = ?", id).Updates(updatedData)
	return result.Error
}

func DeleteFrequency(id uint) error {
	result := configs.DB.Delete(&models.Frequency{}, id)
	return result.Error
}
