package mocks

import "github.com/m5lapp/divesite-monolith/internal/models"

var divePropCave = models.DiveProperty{
	ID:          1,
	Sort:        100,
	IsDefault:   false,
	Name:        "Cave Dive",
	Description: "A dive deep into a cave",
}

var divePropCavern = models.DiveProperty{
	ID:          2,
	Sort:        200,
	IsDefault:   false,
	Name:        "Cavern Dive",
	Description: "A dive into a cave within the reach of natural light",
}

var divePropDecompression = models.DiveProperty{
	ID:          3,
	Sort:        300,
	IsDefault:   false,
	Name:        "Decompression Dive",
	Description: "A dive requiring mandatory decompression stops",
}

type DivePropertyModel struct{}

func (m *DivePropertyModel) AllExist(ids []int) (bool, error) {
	for id := range ids {
		if id < 1 || id > 3 {
			return false, nil
		}
	}

	return true, nil
}

func (m *DivePropertyModel) Exists(id int) (bool, error) {
	return id == 1 || id == 2 || id == 3, nil
}

func (m *DivePropertyModel) GetAllForDive(diveID int) ([]models.DiveProperty, error) {
	if diveID == 1 {
		return []models.DiveProperty{
			divePropCave,
			divePropCavern,
			divePropDecompression,
		}, nil
	}

	return []models.DiveProperty{}, nil
}

func (m *DivePropertyModel) List() ([]models.DiveProperty, error) {
	props := []models.DiveProperty{
		divePropCave,
		divePropCavern,
		divePropDecompression,
	}

	return props, nil
}
