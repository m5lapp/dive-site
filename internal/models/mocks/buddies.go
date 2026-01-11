package mocks

import (
	"time"

	"github.com/m5lapp/divesite-monolith/internal/models"
)

var buddyJohnSmith = models.Buddy{
	ID:              1,
	Version:         1,
	Created:         time.Now(),
	Updated:         time.Now(),
	OwnerID:         1,
	Name:            "John Smith",
	Email:           "john.smith@example.com",
	PhoneNumber:     "07987654321",
	Agency:          &agencyBSAC,
	AgencyMemberNum: "12345",
	DivesWith:       1,
	FirstDiveWith:   &diveDate,
	LastDiveWith:    &diveDate,
	Notes:           "Good diver.",
}

type BuddyModel struct{}

func (m *BuddyModel) Exists(id int) (bool, error) {
	return id == 1, nil
}

func (m *BuddyModel) Insert(
	ownerID int,
	name string,
	emailAddress string,
	phoneNumber string,
	agencyID *int,
	agencyMemberNum string,
	notes string,
) (int, error) {
	return 2, nil
}

func (m *BuddyModel) List(
	userID int,
	pager models.Pager,
	sort []models.SortBuddy,
) ([]models.Buddy, models.PageData, error) {
	pageData := models.PageData{
		FirstPage:    1,
		LastPage:     1,
		CurrentPage:  1,
		PageSize:     20,
		TotalRecords: 1,
	}
	return []models.Buddy{buddyJohnSmith}, pageData, nil
}

func (m *BuddyModel) ListAll(userID int, sort []models.SortBuddy) ([]models.Buddy, error) {
	return []models.Buddy{buddyJohnSmith}, nil
}

var buddyRoleBuddy = models.BuddyRole{
	ID:          1,
	Name:        "Buddy",
	Description: "Dive Buddy",
}

type BuddyRoleModel struct{}

func (m *BuddyRoleModel) Exists(id int) (bool, error) {
	return id == 1, nil
}

func (m *BuddyRoleModel) List() ([]models.BuddyRole, error) {
	return []models.BuddyRole{buddyRoleBuddy}, nil
}
