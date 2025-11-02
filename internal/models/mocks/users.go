package mocks

import (
	"time"

	"github.com/m5lapp/divesite-monolith/internal/models"
)

type UserModel struct{}

func (m *UserModel) Insert(
	name, email, password string,
	divingSince time.Time,
	diveNumberOffset, defaultDivingCountryID int,
	defaultDivingTZ models.TimeZone,
	darkMode bool,
) error {
	switch email {
	case "dupe@example.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	if email == "alice@example.com" && password == "Pa55W0rd" {
		return 1, nil
	}

	return 0, models.ErrInvalidCredentials
}

func (m *UserModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

func (m *UserModel) GetByID(id int) (models.User, error) {
	bangkokTZ, _ := models.NewTimeZone("Asia/Bangkok")

	user := models.User{
		ID:                     1,
		Created:                time.Date(2010, 10, 10, 10, 10, 10, 10, time.UTC),
		Updated:                time.Date(2010, 10, 10, 10, 10, 10, 10, time.UTC),
		Name:                   "Alice Person",
		FriendlyName:           "Alice",
		Email:                  "alice@example.com",
		HashedPassword:         []byte{},
		Suspended:              false,
		Deleted:                false,
		LastLogIn:              time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
		DivingSince:            time.Date(2008, 9, 10, 11, 12, 13, 14, time.UTC),
		DiveNumberOffset:       0,
		DivesLogged:            1,
		TotalDives:             1,
		MaxDiveNumber:          1,
		DefaultDivingCountryID: 17,
		DefaultDivingTZ:        bangkokTZ,
		DarkMode:               false,
	}

	switch id {
	case 1:
		return user, nil
	default:
		return models.User{}, models.ErrNoRecord
	}
}

func (m *UserModel) Update(
	userID int,
	name, email string,
	divingSince time.Time,
	diveNumberOffset, defaultDivingCountryID int,
	defaultDivingTZ models.TimeZone,
	darkMode bool,
) error {
	if userID == 1 {
		if email == "dupe@example.com" {
			return models.ErrDuplicateEmail
		}
		return nil
	}
	return models.ErrNoRecord
}

func (m *UserModel) UpdatePassword(userID int, currentPassword, newPassword string) error {
	if userID == 1 {
		if currentPassword != "Pa55W0rd" {
			return models.ErrInvalidCredentials
		}
		return nil
	}

	return models.ErrNoRecord
}
