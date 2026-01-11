package mocks

import (
	"time"

	"github.com/m5lapp/divesite-monolith/internal/models"
)

var certificationBSACOceanDiver = models.Certification{
	ID:        1,
	Created:   time.Now(),
	Updated:   time.Now(),
	OwnerID:   1,
	Dives:     1,
	FirstDive: &diveDate,
	LastDive:  &diveDate,
	Course:    agencyCourseBSACOceanDiver,
	StartDate: time.Date(2020, time.January, 19, 0, 0, 0, 0, &timeZoneBangkok.Location),
	EndDate:   time.Date(2020, time.January, 22, 0, 0, 0, 0, &timeZoneBangkok.Location),
	Rating:    &ratingSix,
	Operator:  operatorBigBubbles,
	Price:     &price1000AED,
	Notes:     "First introduction to diving.",
}

type CertificationModel struct{}

func (m *CertificationModel) Exists(id int) (bool, error) {
	return id == 1, nil
}

func (m *CertificationModel) Insert(
	ownerID int,
	courseID int,
	startDate time.Time,
	endDate time.Time,
	operatorID int,
	instructorID int,
	priceAmount *float64,
	priceCurrencyID *int,
	rating *int,
	notes string,
) (int, error) {
	return 2, nil
}

func (m *CertificationModel) List(
	userID int,
	pager models.Pager,
	sort []models.SortCert,
) ([]models.Certification, models.PageData, error) {
	pageData := models.PageData{
		FirstPage:    1,
		LastPage:     1,
		CurrentPage:  1,
		PageSize:     20,
		TotalRecords: 1,
	}
	return []models.Certification{certificationBSACOceanDiver}, pageData, nil
}

func (m *CertificationModel) ListAll(
	userID int,
	sort []models.SortCert,
) ([]models.Certification, error) {
	return []models.Certification{certificationBSACOceanDiver}, nil
}
