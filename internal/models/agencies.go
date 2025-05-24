package models

import (
	"context"
	"database/sql"
	"time"
)

type AgencyModelInterface interface {
	List() ([]Agency, error)
}

type Agency struct {
	ID         int
	CommonName string
	FullName   string
	Acronym    string
	URL        string
}

// nullAgency represents an Agency returned from a database that may or may not
// be null.
type nullAgency struct {
	ID         *int
	CommonName *string
	FullName   *string
	Acronym    *string
	URL        *string
}

type AgencyModel struct {
	DB *sql.DB
}

func (na nullAgency) ToAgency() *Agency {
	if na.ID == nil {
		return nil
	}

	return &Agency{
		ID:         *na.ID,
		CommonName: *na.CommonName,
		FullName:   *na.FullName,
		Acronym:    *na.Acronym,
		URL:        *na.URL,
	}
}

// agencyList stores a static, cached slice of agency data so that successive
// requests can bypass the database call.
var agencyList []Agency
var agencyListQuery string = `
    select id, common_name, full_name, acronym, url
      from agencies
  order by common_name
`

func (m *AgencyModel) List() ([]Agency, error) {
	// If the list of agencies has already been populated, then use it.
	if len(agencyList) != 0 {
		return agencyList, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, agencyListQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Agency
	for rows.Next() {
		var record Agency
		err := rows.Scan(
			&record.ID,
			&record.CommonName,
			&record.FullName,
			&record.Acronym,
			&record.URL,
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
	agencyList = records

	return agencyList, nil
}

type AgencyCourse struct {
	ID                int
	Agency            Agency
	Name              string
	URL               string
	IsSpecialtyCourse bool
	IsTechCourse      bool
	IsProCourse       bool
}
