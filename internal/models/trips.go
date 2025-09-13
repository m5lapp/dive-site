package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Price struct {
	Amount   float64
	Currency Currency
}

// nullablePrice represents a Price returned from a database that may or may not
// be null.
type nullablePrice struct {
	Amount   *float64
	Currency nullableCurrency
}

func (np nullablePrice) ToStruct() *Price {
	if np.Amount == nil || np.Currency.ID == nil {
		return nil
	}

	return &Price{
		Amount:   *np.Amount,
		Currency: *np.Currency.ToStruct(),
	}
}

type Trip struct {
	ID          int
	Created     time.Time
	Updated     time.Time
	OwnerID     int
	Name        string
	StartDate   time.Time
	EndDate     time.Time
	Description string
	Rating      *int
	Operator    *Operator
	Price       *Price
	Notes       string
}

func (tr Trip) String() string {
	return fmt.Sprintf(
		"%s (%s to %s)",
		tr.Name,
		tr.StartDate.Format(time.DateOnly),
		tr.EndDate.Format(time.DateOnly),
	)
}

func (tr Trip) Duration() time.Duration {
	return tr.EndDate.Sub(tr.StartDate)
}

type nullableTrip struct {
	ID          *int
	Created     *time.Time
	Updated     *time.Time
	OwnerID     *int
	Name        *string
	StartDate   *time.Time
	EndDate     *time.Time
	Description *string
	Rating      *int
	Operator    nullableOperator
	Price       nullablePrice
	Notes       *string
}

func (nt nullableTrip) ToStruct() *Trip {
	if nt.ID == nil {
		return nil
	}

	return &Trip{
		ID:          *nt.ID,
		Created:     *nt.Created,
		Updated:     *nt.Updated,
		OwnerID:     *nt.OwnerID,
		Name:        *nt.Name,
		StartDate:   *nt.StartDate,
		EndDate:     *nt.EndDate,
		Description: *nt.Description,
		Rating:      nt.Rating,
		Operator:    nt.Operator.ToStruct(),
		Price:       nt.Price.ToStruct(),
		Notes:       *nt.Notes,
	}
}

type TripModelInterface interface {
	Exists(id int) (bool, error)
	Insert(
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
	) (int, error)
	List(userID int, filters ListFilters) ([]Trip, PageData, error)
	ListAll(userID int) ([]Trip, error)
}

var tripSelectQuery string = `
    select count(*) over(),
           tr.id, tr.created_at, tr.updated_at, tr.owner_id, tr.name,
           tr.start_date, tr.end_date, tr.description, tr.rating,
           op.id, op.created_at, op.updated_at, op.owner_id,
           ot.id, ot.name, ot.description,
           op.name, op.street, op.suburb, op.state, op.postcode,
           oc.id, oc.name, oc.iso_number, oc.iso2_code,
           oc.iso3_code, oc.dialing_code, oc.capital,
           ou.id, ou.iso_alpha, ou.iso_number, ou.name, ou.exponent,
           op.website_url, op.email_address, op.phone_number, op.comments,
           tr.price,
           cu.id, cu.iso_alpha, cu.iso_number, cu.name, cu.exponent,
           tr.notes
      from trips tr
 left join operators      op on tr.operator_id = op.id
 left join operator_types ot on op.operator_type_id = ot.id
 left join countries      oc on op.country_id = oc.id
 left join currencies     ou on oc.currency_id = ou.id
 left join currencies     cu on tr.currency_id = cu.id
     where tr.owner_id = $1
`

func tripFromDBRow(rs RowScanner, totalRecords *int, tr *Trip) error {
	op := &nullableOperator{}
	pr := &nullablePrice{}

	err := rs.Scan(
		totalRecords,
		&tr.ID,
		&tr.Created,
		&tr.Updated,
		&tr.OwnerID,
		&tr.Name,
		&tr.StartDate,
		&tr.EndDate,
		&tr.Description,
		&tr.Rating,
		&op.ID,
		&op.Created,
		&op.Updated,
		&op.OwnerID,
		&op.OperatorType.ID,
		&op.OperatorType.Name,
		&op.OperatorType.Description,
		&op.Name,
		&op.Street,
		&op.Suburb,
		&op.State,
		&op.Postcode,
		&op.Country.ID,
		&op.Country.Name,
		&op.Country.ISONumber,
		&op.Country.ISO2Code,
		&op.Country.ISO3Code,
		&op.Country.DialingCode,
		&op.Country.Capital,
		&op.Country.Currency.ID,
		&op.Country.Currency.ISOAlpha,
		&op.Country.Currency.ISONumber,
		&op.Country.Currency.Name,
		&op.Country.Currency.Exponent,
		&op.WebsiteURL,
		&op.EmailAddress,
		&op.PhoneNumber,
		&op.Comments,
		&pr.Amount,
		&pr.Currency.ID,
		&pr.Currency.ISOAlpha,
		&pr.Currency.ISONumber,
		&pr.Currency.Name,
		&pr.Currency.Exponent,
		&tr.Notes,
	)

	if err != nil {
		return err
	}

	tr.Operator = op.ToStruct()
	tr.Price = pr.ToStruct()

	return nil
}

type TripModel struct {
	DB *sql.DB
}

func (m *TripModel) Exists(id int) (bool, error) {
	return idExistsInTable(m.DB, id, "trips", "id")
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
	stmt := `
        insert into trips (
            owner_id, name, start_date, end_date, description, rating,
            operator_id, price, currency_id, notes
        ) values (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
        )
        returning id
    `

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result := m.DB.QueryRowContext(
		ctx,
		stmt,
		ownerID,
		name,
		startDate,
		endDate,
		description,
		rating,
		operatorID,
		priceAmount,
		priceCurrencyID,
		notes,
	)

	var id int
	err := result.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *TripModel) List(userID int, filters ListFilters) ([]Trip, PageData, error) {
	limit := filters.limit()
	offset := filters.offset()
	stmt := fmt.Sprintf("%s limit $2 offset $3", tripSelectQuery)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt, userID, limit, offset)
	if err != nil {
		return nil, PageData{}, err
	}
	defer rows.Close()

	var totalRecords int
	records := []Trip{}
	for rows.Next() {
		var record Trip
		err := tripFromDBRow(rows, &totalRecords, &record)
		if err != nil {
			return nil, PageData{}, err
		}
		records = append(records, record)
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

	return records, paginationData, nil
}

func (m *TripModel) ListAll(userID int) ([]Trip, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, tripSelectQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalRecords int
	var records []Trip
	for rows.Next() {
		var record Trip
		err := tripFromDBRow(rows, &totalRecords, &record)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return records, nil
}
