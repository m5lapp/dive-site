package mocks

import (
	"time"

	"github.com/m5lapp/divesite-monolith/internal/models"
)

var thaiBaht = models.Currency{
	ID:        1,
	ISOAlpha:  "THB",
	ISONumber: 764,
	Name:      "Thai Baht",
	Exponent:  2,
}

var thailand = models.Country{
	ID:          1,
	Name:        "Thailand",
	ISONumber:   764,
	ISO2Code:    "TH",
	ISO3Code:    "THA",
	DialingCode: "66",
	Capital:     "Bangkok",
	Currency:    thaiBaht,
}

var saltWater = models.WaterType{
	ID:          1,
	Name:        "Salt Water",
	Description: "Salt water",
	Density:     1.03,
}

var sea = models.WaterBody{
	ID:          1,
	Name:        "Sea",
	Description: "A sea",
}

var mockDiveSite = models.DiveSite{
	ID:        1,
	Version:   1,
	Created:   time.Now(),
	Updated:   time.Now(),
	OwnerId:   1,
	Name:      "Sail Rock",
	AltName:   "",
	Location:  "Koh Tao",
	Region:    "Surat Thani",
	Country:   thailand,
	TimeZone:  "Asia/Bangkok",
	Latitude:  nil,
	Longitude: nil,
	WaterBody: sea,
	WaterType: saltWater,
	Altitude:  0,
	MaxDepth:  nil,
	Notes:     "Great dive site in the gulf of Thailand.",
	Rating:    nil,
}

type DiveSiteModel struct{}

func (m *DiveSiteModel) Insert(
	ownerId string,
	name string,
	altName string,
	location string,
	region string,
	countryID int,
	timeZone string,
	latitude *float64,
	longitude *float64,
	waterBodyID int,
	waterTypeID int,
	altitude int,
	maxDepth *float64,
	notes string,
	rating *int,
) (int, error) {
	return 2, nil
}

func (m *DiveSiteModel) GetOneByID(id int) (models.DiveSite, error) {
	switch id {
	case 1:
		return mockDiveSite, nil
	default:
		return models.DiveSite{}, models.ErrNoRecord
	}
}

func (m *DiveSiteModel) List(filters map[string]any) ([]models.DiveSite, error) {
	return []models.DiveSite{mockDiveSite}, nil
}
