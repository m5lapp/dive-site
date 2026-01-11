package mocks

import "github.com/m5lapp/divesite-monolith/internal/models"

var currentLight = models.Current{
	StaticDataItem: models.StaticDataItem{
		ID:          3,
		Sort:        30,
		IsDefault:   false,
		Name:        "Light",
		Description: "A noticeable, but weak current",
	},
}

type CurrentModel struct{}

func (m *CurrentModel) Exists(id int) (bool, error) {
	return id == 3, nil
}

func (m *CurrentModel) List(sortByName bool) ([]models.Current, error) {
	return []models.Current{currentLight}, nil
}

var entryPointBoat = models.EntryPoint{
	StaticDataItem: models.StaticDataItem{
		ID:          1,
		Sort:        10,
		IsDefault:   true,
		Name:        "Boat",
		Description: "A dive from a boat",
	},
}

type EntryPointModel struct{}

func (m *EntryPointModel) Exists(id int) (bool, error) {
	return id == 1, nil
}

func (m *EntryPointModel) List(sortByName bool) ([]models.EntryPoint, error) {
	return []models.EntryPoint{entryPointBoat}, nil
}

var gasMixAir = models.GasMix{
	StaticDataItem: models.StaticDataItem{
		ID:          1,
		Sort:        10,
		IsDefault:   true,
		Name:        "Air",
		Description: "Two cylinders mounted on each side",
	},
}

type GasMixModel struct{}

func (m *GasMixModel) Exists(id int) (bool, error) {
	return id == 1, nil
}

func (m *GasMixModel) GetOneByID(id int) (models.GasMix, error) {
	switch id {
	case tankConfigurationSidemount.ID:
		return gasMixAir, nil
	default:
		return models.GasMix{}, models.ErrNoRecord
	}
}

func (m *GasMixModel) List(sortByName bool) ([]models.GasMix, error) {
	return []models.GasMix{gasMixAir}, nil
}

var tankConfigurationSidemount = models.TankConfiguration{
	StaticDataItem: models.StaticDataItem{
		ID:          3,
		Sort:        30,
		IsDefault:   false,
		Name:        "Sidemount",
		Description: "Two cylinders mounted on each side",
	},
	TankCount: 2,
}

type TankConfigurationModel struct{}

func (m *TankConfigurationModel) Exists(id int) (bool, error) {
	return id == 3, nil
}

func (m *TankConfigurationModel) List(sortByName bool) ([]models.TankConfiguration, error) {
	return []models.TankConfiguration{tankConfigurationSidemount}, nil
}

var tankMaterialSteel = models.TankMaterial{
	StaticDataItem: models.StaticDataItem{
		ID:          2,
		Sort:        20,
		IsDefault:   true,
		Name:        "Steel",
		Description: "Commonly used in cooler waters",
	},
}

type TankMaterialModel struct{}

func (m *TankMaterialModel) Exists(id int) (bool, error) {
	return id == 2, nil
}

func (m *TankMaterialModel) List(sortByName bool) ([]models.TankMaterial, error) {
	return []models.TankMaterial{tankMaterialSteel}, nil
}

var wavesModerate = models.Waves{
	StaticDataItem: models.StaticDataItem{
		ID:          4,
		Sort:        40,
		IsDefault:   false,
		Name:        "moderate",
		Description: "Small waves with breaking crests, fairly frequent whitecap",
	},
}

type WavesModel struct{}

func (m *WavesModel) Exists(id int) (bool, error) {
	return id == 4, nil
}

func (m *WavesModel) List(sortByName bool) ([]models.Waves, error) {
	return []models.Waves{wavesModerate}, nil
}
