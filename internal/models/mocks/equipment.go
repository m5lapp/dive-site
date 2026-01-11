package mocks

import "github.com/m5lapp/divesite-monolith/internal/models"

var equipmentBoots3mm = models.Equipment{
	ID:          1,
	Sort:        10,
	IsDefault:   false,
	Name:        "Boots (3mm)",
	Description: "3mm Wetsuit boots",
}

var equipmentBoots5mm = models.Equipment{
	ID:          2,
	Sort:        20,
	IsDefault:   false,
	Name:        "Boots (5mm)",
	Description: "5mm Wetsuit boots",
}

var equipmentBoots7mm = models.Equipment{
	ID:          3,
	Sort:        30,
	IsDefault:   false,
	Name:        "Boots (7mm)",
	Description: "7mm Wetsuit boots",
}

type EquipmentModel struct{}

func (m *EquipmentModel) AllExist(ids []int) (bool, error) {
	for id := range ids {
		if id < 1 || id > 3 {
			return false, nil
		}
	}

	return true, nil
}

func (m *EquipmentModel) Exists(id int) (bool, error) {
	return id == 1 || id == 2 || id == 3, nil
}

func (m *EquipmentModel) GetAllForDive(diveID int) ([]models.Equipment, error) {
	if diveID == 1 {
		return []models.Equipment{
			equipmentBoots3mm,
			equipmentBoots5mm,
			equipmentBoots7mm,
		}, nil
	}

	return []models.Equipment{}, nil
}

func (m *EquipmentModel) List() ([]models.Equipment, error) {
	props := []models.Equipment{
		equipmentBoots3mm,
		equipmentBoots5mm,
		equipmentBoots7mm,
	}

	return props, nil
}
