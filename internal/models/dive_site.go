package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

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

type WaterType struct {
	ID          int
	Name        string
	Description string
	Density     float64
}

type WaterBody struct {
	ID          int
	Name        string
	Description string
}

type DiveSite struct {
	ID        int
	Version   int
	Created   time.Time
	Updated   time.Time
	OwnerId   string
	Name      string
	AltName   string
	Location  string
	Region    string
	Country   Country
	TimeZone  string
	Latitude  *float64
	Longitude *float64
	WaterBody WaterBody
	WaterType WaterType
	Altitude  int
	MaxDepth  *float64
	Notes     string
	Rating    *int
}

type DiveSiteModelInterface interface {
	Insert(
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
	) (int, error)

	GetOneByID(id int) (DiveSite, error)

	List(ListControls ListFilters) ([]DiveSite, PageData, error)
}

var diveSiteSelectQuery string = `
    select
            count(*) over(), ds.id, ds.version, ds.created_at, ds.updated_at,
            ds.owner_id, ds.name, ds.alt_name, ds.location, ds.region,
            ds.timezone, ds.latitude, ds.longitude, ds.altitude, ds.max_depth,
            ds.notes, ds.rating, co.id, co.name, co.iso_number, co.iso2_code,
            co.iso3_code, co.dialing_code, co.capital, cu.id, cu.iso_alpha,
            cu.iso_number, cu.name, cu.exponent, wb.id, wb.name, wb.description,
            wt.id, wt.name, wt.description, wt.density
      from dive_sites ds
 left join countries    co on ds.country_id = co.id
 left join currencies   cu on co.currency_id = cu.id
 left join water_bodies wb on ds.water_body_id = wb.id
 left join water_types  wt on ds.water_type_id = wt.id
     where ds.owner_id = $1
`

func diveSiteFromDBRow(rs RowScanner, totalRecords *int, ds *DiveSite) error {
	return rs.Scan(
		totalRecords,
		&ds.ID,
		&ds.Version,
		&ds.Created,
		&ds.Updated,
		&ds.OwnerId,
		&ds.Name,
		&ds.AltName,
		&ds.Location,
		&ds.Region,
		&ds.TimeZone,
		&ds.Latitude,
		&ds.Longitude,
		&ds.Altitude,
		&ds.MaxDepth,
		&ds.Notes,
		&ds.Rating,
		&ds.Country.ID,
		&ds.Country.Name,
		&ds.Country.ISONumber,
		&ds.Country.ISO2Code,
		&ds.Country.ISO3Code,
		&ds.Country.DialingCode,
		&ds.Country.Capital,
		&ds.Country.Currency.ID,
		&ds.Country.Currency.ISOAlpha,
		&ds.Country.Currency.ISONumber,
		&ds.Country.Currency.Name,
		&ds.Country.Currency.Exponent,
		&ds.WaterBody.ID,
		&ds.WaterBody.Name,
		&ds.WaterBody.Description,
		&ds.WaterType.ID,
		&ds.WaterType.Name,
		&ds.WaterType.Description,
		&ds.WaterType.Density,
	)
}

type DiveSiteModel struct {
	DB *sql.DB
}

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
	stmt := `
        insert into dive_sites (
            owner_id, name, alt_name, location, region, country_id, timezone,
            latitude, longitude, water_body_id, water_type_id, altitude,
            max_depth, notes, rating
        ) values (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
        )
        returning id
    `

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result := m.DB.QueryRowContext(
		ctx,
		stmt,
		ownerId,
		name,
		altName,
		location,
		region,
		countryID,
		timeZone,
		latitude,
		longitude,
		waterBodyID,
		waterTypeID,
		altitude,
		maxDepth,
		notes,
		rating,
	)

	var id int
	err := result.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *DiveSiteModel) GetOneByID(id int) (DiveSite, error) {
	stmt := fmt.Sprintf("%s and ds.id = $2", diveSiteSelectQuery)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var totalRecords int
	var diveSite DiveSite
	row := m.DB.QueryRowContext(ctx, stmt, "abc123", id)
	err := diveSiteFromDBRow(row, &totalRecords, &diveSite)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return DiveSite{}, ErrNoRecord
		} else {
			return DiveSite{}, err
		}
	}

	return diveSite, nil
}

func (m *DiveSiteModel) List(filters ListFilters) ([]DiveSite, PageData, error) {
	limit := filters.limit()
	offset := filters.offset()
	stmt := fmt.Sprintf("%s limit $2 offset $3", diveSiteSelectQuery)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt, "abc123", limit, offset)
	if err != nil {
		return nil, PageData{}, err
	}
	defer rows.Close()

	var totalRecords int
	var diveSites []DiveSite
	for rows.Next() {
		var ds DiveSite
		err := diveSiteFromDBRow(rows, &totalRecords, &ds)
		if err != nil {
			return nil, PageData{}, err
		}
		diveSites = append(diveSites, ds)
	}

	err = rows.Err()
	if err != nil {
		return nil, PageData{}, err
	}

	paginationData := newPaginationData(
		totalRecords,
		filters.page,
		filters.pageSize,
	)

	return diveSites, paginationData, nil
}
