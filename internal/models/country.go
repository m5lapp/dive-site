package models

import (
	"context"
	"database/sql"
	"time"
)

type CurrencyModel struct {
	DB *sql.DB
}

type CurrencyModelInterface interface {
	List() ([]Currency, error)
}

type Currency struct {
	ID        int
	ISOAlpha  string
	ISONumber int
	Name      string
	Exponent  int
}

// nullableCurrency represents a Currency returned from a database that may or
// may not be null.
type nullableCurrency struct {
	ID        *int
	ISOAlpha  *string
	ISONumber *int
	Name      *string
	Exponent  *int
}

func (nc nullableCurrency) ToStruct() *Currency {
	if nc.ID == nil {
		return nil
	}

	return &Currency{
		ID:        *nc.ID,
		ISOAlpha:  *nc.Name,
		ISONumber: *nc.ISONumber,
		Name:      *nc.Name,
		Exponent:  *nc.Exponent,
	}
}

// currencyList stores a static, cached slice of Currency data so that
// successive requests can bypass the database call.
var currencyList []Currency

var currencyListQuery string = `
     select cu.id, cu.iso_alpha, cu.iso_number, cu.name, cu.exponent
       from currencies cu
   order by cu.name
`

func (m *CurrencyModel) List() ([]Currency, error) {
	// If the list of currencies has already been populated, then use it.
	if len(currencyList) != 0 {
		return currencyList, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, currencyListQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Currency
	for rows.Next() {
		var record Currency
		err := rows.Scan(
			&record.ID,
			&record.ISOAlpha,
			&record.ISONumber,
			&record.Name,
			&record.Exponent,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// Cache the response for faster future calls.
	currencyList = records

	return records, nil
}

type CountryModel struct {
	DB *sql.DB
}

type CountryModelInterface interface {
	List() ([]Country, error)
}

type Country struct {
	ID          int
	Name        string
	ISONumber   int
	ISO2Code    string
	ISO3Code    string
	DialingCode string
	Capital     string
	Currency    Currency
}

// nullableCountry represents a Country returned from a database that may or may
// not be null.
type nullableCountry struct {
	ID          *int
	Name        *string
	ISONumber   *int
	ISO2Code    *string
	ISO3Code    *string
	DialingCode *string
	Capital     *string
	Currency    nullableCurrency
}

func (nc nullableCountry) ToStruct() *Country {
	if nc.ID == nil || nc.Currency.ID == nil {
		return nil
	}

	return &Country{
		ID:          *nc.ID,
		Name:        *nc.Name,
		ISONumber:   *nc.ISONumber,
		ISO2Code:    *nc.ISO2Code,
		ISO3Code:    *nc.ISO3Code,
		DialingCode: *nc.DialingCode,
		Capital:     *nc.Capital,
		Currency:    *nc.Currency.ToStruct(),
	}
}

// countryList stores a static, cached slice of Country data so that successive
// requests can bypass the database call.
var countryList []Country

var countryListQuery string = `
     select co.id, co.name, co.iso_number, co.iso2_code, co.iso3_code,
            co.dialing_code, co.capital, cu.id, cu.iso_alpha, cu.iso_number,
            cu.name, cu.exponent
       from countries co
  left join currencies cu on co.currency_id = cu.id
   order by co.name
`

func (m *CountryModel) List() ([]Country, error) {
	// If the list of countries has already been populated, then use it.
	if len(countryList) != 0 {
		return countryList, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, countryListQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []Country
	for rows.Next() {
		var country Country
		err := rows.Scan(
			&country.ID,
			&country.Name,
			&country.ISONumber,
			&country.ISO2Code,
			&country.ISO3Code,
			&country.DialingCode,
			&country.Capital,
			&country.Currency.ID,
			&country.Currency.ISOAlpha,
			&country.Currency.ISONumber,
			&country.Currency.Name,
			&country.Currency.Exponent,
		)
		if err != nil {
			return nil, err
		}
		countries = append(countries, country)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// Cache the response for faster future calls.
	countryList = countries

	return countries, nil
}
