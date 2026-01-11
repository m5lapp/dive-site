package mocks

import (
	"time"

	"github.com/m5lapp/divesite-monolith/internal/models"
)

var currencyThaiBaht = models.Currency{
	ID:        1,
	ISOAlpha:  "THB",
	ISONumber: 764,
	Name:      "Thai Baht",
	Exponent:  2,
}

var countryThailand = models.Country{
	ID:          1,
	Name:        "Thailand",
	ISONumber:   764,
	ISO2Code:    "TH",
	ISO3Code:    "THA",
	DialingCode: "66",
	Capital:     "Bangkok",
	Currency:    currencyThaiBaht,
}

var timeZoneBangkok, _ = models.NewTimeZone("Asia/Bangkok")

var waterTypeSaltWater = models.WaterType{
	ID:          1,
	Name:        "Salt Water",
	Description: "Salt water",
}

var waterBodySea = models.WaterBody{
	ID:          1,
	Name:        "Sea",
	Description: "A sea",
}

var diveSiteSailRock = models.DiveSite{
	ID:        1,
	Version:   1,
	Created:   time.Now(),
	Updated:   time.Now(),
	OwnerId:   1,
	Name:      "Sail Rock",
	AltName:   "",
	Location:  "Koh Tao",
	Region:    "Surat Thani",
	Country:   countryThailand,
	TimeZone:  timeZoneBangkok,
	Latitude:  nil,
	Longitude: nil,
	WaterBody: waterBodySea,
	WaterType: waterTypeSaltWater,
	Altitude:  0,
	MaxDepth:  nil,
	Notes:     "Great dive site in the gulf of Thailand.",
	Rating:    nil,
}

type DiveSiteModel struct{}

func (m *DiveSiteModel) Insert(
	ownerId int,
	name string,
	altName string,
	location string,
	region string,
	countryID int,
	timeZone models.TimeZone,
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

func (m *DiveSiteModel) GetOneByID(id, userID int) (models.DiveSite, error) {
	switch id {
	case 1:
		return diveSiteSailRock, nil
	default:
		return models.DiveSite{}, models.ErrNoRecord
	}
}

func (m *DiveSiteModel) List(
	diverID int,
	filters models.Pager,
	sort []models.SortDiveSite,
) ([]models.DiveSite, models.PageData, error) {
	pageData := models.PageData{
		FirstPage:    1,
		LastPage:     1,
		CurrentPage:  1,
		PageSize:     1,
		TotalRecords: 1,
	}
	return []models.DiveSite{diveSiteSailRock}, pageData, nil
}

func (m *DiveSiteModel) ListAll(diverID int) ([]models.DiveSite, error) {
	return []models.DiveSite{diveSiteSailRock}, nil
}

func (m *DiveSiteModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

func (m *DiveSiteModel) Update(
	id int,
	version int,
	name string,
	altName string,
	location string,
	region string,
	countryID int,
	timeZone models.TimeZone,
	latitude *float64,
	longitude *float64,
	waterBodyID int,
	waterTypeID int,
	altitude int,
	maxDepth *float64,
	notes string,
	rating *int,
) error {
	if id == 1 {
		if version == 2 {
			return models.ErrUpdateConflict
		}

		return nil
	}

	return models.ErrNoRecord
}

type WaterBodyModel struct{}

func (m *WaterBodyModel) List() ([]models.WaterBody, error) {
	return []models.WaterBody{waterBodySea}, nil
}

type WaterTypeModel struct{}

func (m *WaterTypeModel) List() ([]models.WaterType, error) {
	return []models.WaterType{waterTypeSaltWater}, nil
}
