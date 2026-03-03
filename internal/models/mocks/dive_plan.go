package mocks

import (
	"time"

	"github.com/m5lapp/diveplanner"
	"github.com/m5lapp/diveplanner/gasmix"
	"github.com/m5lapp/divesite-monolith/internal/models"
)

var planGasMixAir = gasmix.NewAirMix()

var divePlan28mStops = []*diveplanner.DivePlanStop{
	&diveplanner.DivePlanStop{Depth: 28.0, Duration: 8},
	&diveplanner.DivePlanStop{Depth: 15.0, Duration: 10},
	&diveplanner.DivePlanStop{Depth: 8.0, Duration: 5},
	&diveplanner.DivePlanStop{Depth: 5.0, Duration: 3, Comment: "Safety stop"},
}

var divePlan28m = models.DivePlan{
	ID:      1,
	Version: 1,
	OwnerId: 1,
	DivePlan: diveplanner.DivePlan{
		Created:         time.Now(),
		Updated:         time.Now(),
		Name:            "test Plan",
		Notes:           "Good, conservative dive plan.",
		IsSoloDive:      false,
		DescentRate:     18.0,
		AscentRate:      9.0,
		SACRate:         11.0,
		TankCount:       1,
		TankCapacity:    11,
		WorkingPressure: 200,
		DiveFactor:      1.2,
		GasMix:          planGasMixAir,
		MaxPPO2:         1.4,
		Stops:           divePlan28mStops,
	},
}

type DivePlanModel struct{}

func (m *DivePlanModel) Insert(
	ownerID int,
	name string,
	notes string,
	isSoloDive bool,
	descentRate float64,
	ascentRate float64,
	sacRate float64,
	tankCount int,
	tankVolume float64,
	workingPressure int,
	diveFactor float64,
	fn2 float64,
	fhe float64,
	maxPPO2 float64,
	stops []models.DivePlanStopInput,
) (int, error) {
	return 2, nil
}

func (m *DivePlanModel) Update(
	id int,
	ownerID int,
	name string,
	notes string,
	isSoloDive bool,
	descentRate float64,
	ascentRate float64,
	sacRate float64,
	tankCount int,
	tankVolume float64,
	workingPressure int,
	diveFactor float64,
	fn2 float64,
	fhe float64,
	maxPPO2 float64,
	stops []models.DivePlanStopInput,
) error {
	if id == 1 {
		return nil
	}

	return models.ErrNoRecord
}

func (m *DivePlanModel) GetOneByID(id, userID int) (models.DivePlan, error) {
	switch id {
	case 1:
		return divePlan28m, nil
	default:
		return models.DivePlan{}, models.ErrNoRecord
	}
}

func (m *DivePlanModel) List(
	diverID int,
	filters models.Pager,
	sort []models.SortDivePlan,
) ([]models.DivePlan, models.PageData, error) {
	pageData := models.PageData{
		FirstPage:    1,
		LastPage:     1,
		CurrentPage:  1,
		PageSize:     1,
		TotalRecords: 1,
	}
	return []models.DivePlan{divePlan28m}, pageData, nil
}

func (m *DivePlanModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}
