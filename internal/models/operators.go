package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type OperatorType struct {
	ID          int
	Name        string
	Description string
}

type Operator struct {
	ID           int
	Created      time.Time
	Updated      time.Time
	OwnerID      int
	OperatorType OperatorType
	Name         string
	Street       string
	Suburb       string
	State        string
	Postcode     string
	Country      Country
	WebsiteURL   string
	EmailAddress string
	PhoneNumber  string
	Comments     string
}

type OperatorModelInterface interface {
	Insert(
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
	) (int, error)
	List(filters ListFilters) ([]Operator, PageData, error)
	ListAll() ([]Operator, error)
}

var operatorSelectQuery string = `
    select count(*) over(), op.id, op.created_at, op.updated_at, op.owner_id,
           ot.id, ot.name, ot.description,
           op.name, op.street, op.suburb, op.state, op.postcode,
           co.id, co.name, co.iso_number, co.iso2_code,
           co.iso3_code, co.dialing_code, co.capital,
           cu.id, cu.iso_alpha, cu.iso_number, cu.name, cu.exponent,
           op.website_url, op.email_address, op.phone_number, op.comments
      from operators op
 left join operator_types ot on op.operator_type_id = ot.id
 left join countries      co on op.country_id = co.id
 left join currencies     cu on co.currency_id = cu.id
`

func operatorFromDBRow(rs RowScanner, totalRecords *int, op *Operator) error {
	return rs.Scan(
		totalRecords,
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
	)
}

type OperatorModel struct {
	DB *sql.DB
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
	stmt := `
        insert into operators (
            owner_id, operator_type_id, name, street, suburb, state, postcode,
            country_id, website_url, email_address, phone_number, comments
        ) values (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
        )
        returning id
    `

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result := m.DB.QueryRowContext(
		ctx,
		stmt,
		ownerID,
		operatorTypeID,
		name,
		street,
		suburb,
		state,
		postcode,
		countryID,
		websiteURL,
		emailAddress,
		phoneNumber,
		comments,
	)

	var id int
	err := result.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *OperatorModel) List(filters ListFilters) ([]Operator, PageData, error) {
	limit := filters.limit()
	offset := filters.offset()
	stmt := fmt.Sprintf("%s limit $1 offset $2", operatorSelectQuery)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt, limit, offset)
	if err != nil {
		return nil, PageData{}, err
	}
	defer rows.Close()

	var totalRecords int
	operators := []Operator{}
	for rows.Next() {
		var op Operator
		err := operatorFromDBRow(rows, &totalRecords, &op)
		if err != nil {
			return nil, PageData{}, err
		}
		operators = append(operators, op)
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

	return operators, paginationData, nil
}

func (m *OperatorModel) ListAll() ([]Operator, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, operatorSelectQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalRecords int
	var operators []Operator
	for rows.Next() {
		var op Operator
		err := operatorFromDBRow(rows, &totalRecords, &op)
		if err != nil {
			return nil, err
		}
		operators = append(operators, op)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return operators, nil
}

type OperatorTypeModel struct {
	DB *sql.DB
}

type OperatorTypeModelInterface interface {
	List() ([]OperatorType, error)
}

// operatorTypeList stores a static, cached slice of OperatorType data so that
// successive requests can bypass the database call.
var operatorTypeList []OperatorType
var operatorTypeListQuery string = `
    select id, name, description
      from operator_types 
  order by name
`

func (m *OperatorTypeModel) List() ([]OperatorType, error) {
	// If the list of operator types has already been populated, then use it.
	if len(operatorTypeList) != 0 {
		return operatorTypeList, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, operatorTypeListQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var operatorTypes []OperatorType
	for rows.Next() {
		var operatorType OperatorType
		err := rows.Scan(&operatorType.ID, &operatorType.Name, &operatorType.Description)
		if err != nil {
			return nil, err
		}
		operatorTypes = append(operatorTypes, operatorType)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// Cache the response for faster future calls.
	operatorTypeList = operatorTypes

	return operatorTypes, nil
}
