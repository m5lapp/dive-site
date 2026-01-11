package mocks

import "github.com/m5lapp/divesite-monolith/internal/models"

var currencyAED = models.Currency{
	ID:        1,
	ISOAlpha:  "AED",
	ISONumber: 784,
	Name:      "United Arab Emirates Dirham",
	Exponent:  2,
}

var price1000AED = models.Price{
	Amount:   1000.00,
	Currency: currencyAED,
}

type CurrencyModel struct{}

func (m *CurrencyModel) List() ([]models.Currency, error) {
	return []models.Currency{currencyAED}, nil
}

var countryAFG = models.Country{
	ID:          1,
	Name:        "Afghanistan",
	ISONumber:   4,
	ISO2Code:    "AF",
	ISO3Code:    "AFG",
	DialingCode: "93",
	Capital:     "Kabul",
	Currency:    currencyAED,
}

type CountryModel struct{}

func (m *CountryModel) List() ([]models.Country, error) {
	return []models.Country{countryAFG}, nil
}
