package models

import (
	"context"
	"database/sql"
	"fmt"
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

// nullableAgency represents an Agency returned from a database that may or may
// not be null.
type nullableAgency struct {
	ID         *int
	CommonName *string
	FullName   *string
	Acronym    *string
	URL        *string
}

type AgencyModel struct {
	DB *sql.DB
}

func (na nullableAgency) ToStruct() *Agency {
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

type AgencyCourseModelInterface interface {
	List() ([]AgencyCourse, error)
}

type AgencyCourseModel struct {
	DB *sql.DB
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

func (ac AgencyCourse) String() string {
	return fmt.Sprintf("%s - %s", ac.Agency.CommonName, ac.Name)
}

type nullableAgencyCourse struct {
	ID                *int
	Agency            nullableAgency
	Name              *string
	URL               *string
	IsSpecialtyCourse *bool
	IsTechCourse      *bool
	IsProCourse       *bool
}

func (na nullableAgencyCourse) ToStruct() *AgencyCourse {
	if na.ID == nil {
		return nil
	}

	return &AgencyCourse{
		ID:                *na.ID,
		Agency:            *na.Agency.ToStruct(),
		Name:              *na.Name,
		URL:               *na.URL,
		IsSpecialtyCourse: *na.IsSpecialtyCourse,
		IsTechCourse:      *na.IsTechCourse,
		IsProCourse:       *na.IsProCourse,
	}
}

// agencyCourseList stores a static, cached slice of agency data so that
// successive requests can bypass the database call.
var agencyCourseList []AgencyCourse
var agencyCourseListQuery string = `
    select ac.id,
           ac.agency_id, ag.common_name, ag.full_name, ag.acronym, ag.url,
           ac.name, ac.url, ac.is_specialty_course, ac.is_tech_course,
           ac.is_pro_course
      from agency_courses ac
inner join agencies ag on ac.agency_id = ag.id
  order by ag.common_name, ac.name
`

func (m *AgencyCourseModel) List() ([]AgencyCourse, error) {
	// If the list of agency courses has already been populated, then use it.
	if len(agencyCourseList) != 0 {
		return agencyCourseList, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, agencyCourseListQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []AgencyCourse
	for rows.Next() {
		var record AgencyCourse
		err := rows.Scan(
			&record.ID,
			&record.Agency.ID,
			&record.Agency.CommonName,
			&record.Agency.FullName,
			&record.Agency.Acronym,
			&record.Agency.URL,
			&record.Name,
			&record.URL,
			&record.IsSpecialtyCourse,
			&record.IsTechCourse,
			&record.IsProCourse,
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
	agencyCourseList = records

	return agencyCourseList, nil
}
