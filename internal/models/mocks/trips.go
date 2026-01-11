package mocks

import (
	"time"

	"github.com/m5lapp/divesite-monolith/internal/models"
)

var ratingSix int = 6

var tripLiveaboard = models.Trip{
	ID:          1,
	Created:     time.Now(),
	Updated:     time.Now(),
	OwnerID:     1,
	Dives:       1,
	FirstDive:   &diveDate,
	LastDive:    &diveDate,
	Name:        "Big Splash Liveaboard",
	StartDate:   time.Date(2020, time.January, 17, 0, 0, 0, 0, &timeZoneBangkok.Location),
	EndDate:     time.Date(2020, time.January, 24, 0, 0, 0, 0, &timeZoneBangkok.Location),
	Description: "Big Bubbles Liveaboard",
	Rating:      &ratingSix,
	Operator:    &operatorBigBubbles,
	Price:       &price1000AED,
	Notes:       "Good, fun liveaboard with lots of sharks.",
}

type TripModel struct{}

func (m *TripModel) Exists(id int) (bool, error) {
	return id == 1, nil
}

func (m *TripModel) Insert(
	ownerID int,
	name string,
	startDate time.Time,
	endDate time.Time,
	description string,
	rating *int,
	operatorID *int,
	priceAmount *float64,
	priceCurrencyID *int,
	notes string,
) (int, error) {
	return 2, nil
}

func (m *TripModel) List(
	userID int,
	pager models.Pager,
	sort []models.SortTrip,
) ([]models.Trip, models.PageData, error) {
	pageData := models.PageData{
		FirstPage:    1,
		LastPage:     1,
		CurrentPage:  1,
		PageSize:     20,
		TotalRecords: 1,
	}
	return []models.Trip{tripLiveaboard}, pageData, nil
}

func (m *TripModel) ListAll(userID int, sort []models.SortTrip) ([]models.Trip, error) {
	return []models.Trip{tripLiveaboard}, nil
}
