package mocks

import (
	"time"

	"github.com/m5lapp/divesite-monolith/internal/models"
)

var (
	diveDate = time.Date(
		2020,
		time.January,
		19,
		14,
		21,
		0,
		0,
		&timeZoneBangkok.Location,
	)
	avgDepth12      float64       = 12.0
	safetyStop3Mins time.Duration = 3 * time.Minute
	waterTemp28     int           = 28
	airTemp31       int           = 31
	visibility15m   float64       = 15.0
	weight8kg       float64       = 8.0
	pressureIn210   int           = 210
	pressureOut65   int           = 65
)

var dive1 = models.Dive{
	ID:                1,
	Version:           1,
	Created:           time.Now(),
	Updated:           time.Now(),
	OwnerID:           1,
	Number:            1,
	Activity:          "Fun Dive",
	DiveSite:          diveSiteSailRock,
	Operator:          &operatorBigBubbles,
	Price:             &price1000AED,
	Trip:              &tripLiveaboard,
	Certification:     &certificationBSACOceanDiver,
	DateTimeIn:        diveDate,
	SurfaceInterval:   nil,
	MaxDepth:          17.6,
	AvgDepth:          &avgDepth12,
	BottomTime:        45 * time.Second,
	SafetyStop:        &safetyStop3Mins,
	WaterTemp:         &waterTemp28,
	AirTemp:           &airTemp31,
	Visibility:        &visibility15m,
	Current:           &currentLight,
	Waves:             &wavesModerate,
	Buddy:             &buddyJohnSmith,
	BuddyRole:         &buddyRoleBuddy,
	Weight:            &weight8kg,
	WeightNotes:       "",
	Equipment:         []models.Equipment{equipmentBoots5mm},
	EquipmentNotes:    "Wore my new 5mm boots.",
	TankConfiguration: tankConfigurationSidemount,
	TankMaterial:      tankMaterialSteel,
	TankVolume:        11.2,
	GasMix:            gasMixAir,
	FO2:               0.21,
	PressureIn:        &pressureIn210,
	PressureOut:       &pressureOut65,
	GasMixNotes:       "",
	EntryPoint:        entryPointBoat,
	Properties:        []models.DiveProperty{divePropCavern},
	Rating:            &ratingSix,
	Notes:             "Great first dive.",
}

type DiveModel struct{}

func (m *DiveModel) GetDiveStats(userID int) (models.DiveStats, error) {
	return models.DiveStats{}, nil
}

func (m *DiveModel) GetOneByID(ownerID, id int) (models.Dive, error) {
	if ownerID == 1 && id == 1 {
		return dive1, nil
	}

	return models.Dive{}, models.ErrNoRecord
}

func (m *DiveModel) Insert(
	ownerID int,
	number int,
	activity string,
	diveSiteID int,
	operatorID *int,
	priceAmount *float64,
	priceCurrencyID *int,
	tripID *int,
	certificationID *int,
	dateTimeIn time.Time,
	maxDepth float64,
	avgDepth *float64,
	bottomTime time.Duration,
	safetyStop *time.Duration,
	waterTemp *int,
	airTemp *int,
	visibility *float64,
	currentID *int,
	wavesID *int,
	buddyID *int,
	buddyRoleID *int,
	weight *float64,
	weightNotes string,
	equipmentIDs []int,
	equipmentNotes string,
	tankConfigurationID int,
	tankMaterialID int,
	tankVolume float64,
	gasMixID int,
	fo2 float64,
	pressureIn *int,
	pressureOut *int,
	gasMixNotes string,
	entryPointID int,
	propertyIDs []int,
	rating *int,
	notes string,
) (int, error) {
	return 2, nil
}

func (m *DiveModel) Update(
	id int,
	ownerID int,
	number int,
	activity string,
	diveSiteID int,
	operatorID *int,
	priceAmount *float64,
	priceCurrencyID *int,
	tripID *int,
	certificationID *int,
	dateTimeIn time.Time,
	maxDepth float64,
	avgDepth *float64,
	bottomTime time.Duration,
	safetyStop *time.Duration,
	waterTemp *int,
	airTemp *int,
	visibility *float64,
	currentID *int,
	wavesID *int,
	buddyID *int,
	buddyRoleID *int,
	weight *float64,
	weightNotes string,
	equipmentIDs []int,
	equipmentNotes string,
	tankConfigurationID int,
	tankMaterialID int,
	tankVolume float64,
	gasMixID int,
	fo2 float64,
	pressureIn *int,
	pressureOut *int,
	gasMixNotes string,
	entryPointID int,
	propertyIDs []int,
	rating *int,
	notes string,
) error {
	return nil
}

func (m *DiveModel) List(
	userID int,
	pager models.Pager,
	filter models.DiveFilter,
	sort []models.SortDive,
) ([]models.Dive, models.PageData, error) {
	switch userID {
	case 1:
		return []models.Dive{dive1}, models.PageData{}, nil
	default:
		return []models.Dive{}, models.PageData{}, nil
	}
}
