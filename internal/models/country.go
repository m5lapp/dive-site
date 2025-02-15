package models

import (
	"context"
	"database/sql"
	"time"
)

type CountryModel struct {
	DB *sql.DB
}

type CountryModelInterface interface {
	List() ([]Country, error)
}

type Currency struct {
	ID        int
	ISOAlpha  string
	ISONumber int
	Name      string
	Exponent  int
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
