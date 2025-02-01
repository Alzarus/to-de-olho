package repositories

import (
	"to-de-olho-api/configs"
	"to-de-olho-api/models"
)

func CreateContract(contract *models.Contract) error {
	return configs.DB.Create(contract).Error
}

func CreateContracts(contracts []models.Contract) error {
	return configs.DB.Create(&contracts).Error
}

func GetAllContracts() ([]models.Contract, error) {
	var contracts []models.Contract
	err := configs.DB.Find(&contracts).Error
	return contracts, err
}

func GetContractByID(id uint) (models.Contract, error) {
	var contract models.Contract
	err := configs.DB.First(&contract, id).Error
	return contract, err
}

func GetLatestContract() (models.Contract, error) {
	var contract models.Contract
	err := configs.DB.Order("data_publicacao DESC").First(&contract).Error
	return contract, err
}

func UpdateContract(id uint, updatedData *models.Contract) error {
	return configs.DB.Model(&models.Contract{}).Where("id = ?", id).Updates(updatedData).Error
}

func DeleteContract(id uint) error {
	return configs.DB.Delete(&models.Contract{}, id).Error
}
