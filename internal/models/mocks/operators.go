package mocks

import (
	"time"

	"github.com/m5lapp/divesite-monolith/internal/models"
)

var operatorTypeDiveClub = models.OperatorType{
	ID:          1,
	Name:        "Dive Club",
	Description: "An organised dive club",
}

type OperatorTypeModel struct{}

func (m *OperatorTypeModel) Exists(id int) (bool, error) {
	return id == 1, nil
}

func (m *OperatorTypeModel) List() ([]models.OperatorType, error) {
	return []models.OperatorType{operatorTypeDiveClub}, nil
}

var operatorBigBubbles = models.Operator{
	ID:           1,
	Created:      time.Now(),
	Updated:      time.Now(),
	OwnerID:      1,
	Dives:        1,
	FirstDive:    &diveDate,
	LastDive:     &diveDate,
	OperatorType: operatorTypeDiveClub,
	Name:         "Big Bubbles",
	Street:       "123, Fake Street",
	Suburb:       "Bala Murghab",
	State:        "Badghis",
	Postcode:     "90210",
	Country:      countryAFG,
	WebsiteURL:   "https://big-bubbles.com/",
	EmailAddress: "info@big-bubbles.com",
	PhoneNumber:  "",
	Comments:     "",
}

type OperatorModel struct{}

func (m *OperatorModel) Exists(id int) (bool, error) {
	return id == 1, nil
}

func (m *OperatorModel) Insert(
	ownerID int,
	operatorTypeID int,
	name string,
	street string,
	suburb string,
	state string,
	postcode string,
	countryID int,
	websiteURL string,
	emailAddress string,
	phoneNumber string,
	comments string,
) (int, error) {
	return 2, nil
}

func (m *OperatorModel) List(
	userID int,
	pager models.Pager,
	sort []models.SortOperator,
) ([]models.Operator, models.PageData, error) {
	pageData := models.PageData{
		FirstPage:    1,
		LastPage:     1,
		CurrentPage:  1,
		PageSize:     20,
		TotalRecords: 1,
	}
	return []models.Operator{operatorBigBubbles}, pageData, nil
}

func (m *OperatorModel) ListAll(userID int, sort []models.SortOperator) ([]models.Operator, error) {
	return []models.Operator{operatorBigBubbles}, nil
}
