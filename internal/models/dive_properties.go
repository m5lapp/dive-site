package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type DivePropertyModel struct {
	DB *sql.DB
}

type DivePropertyModelInterface interface {
	AllExist(ids []int) (bool, error)
	Exists(id int) (bool, error)
	List() ([]DiveProperty, error)
}

type DiveProperty struct {
	ID          int
	Sort        int
	IsDefault   bool
	Name        string
	Description string
}

func (e DiveProperty) String() string {
	return fmt.Sprintf("%s - %s", e.Name, e.Description)
}

// divePropertyList stores a static, cached slice of DiveProperty data so that
// successive requests can bypass the database call.
var divePropertyList []DiveProperty

var divePropertySelectQuery string = `
     select dp.id, dp.sort, dp.is_default, dp.name, dp.description
       from dive_properties dp
   order by dp.name
`

func (m *DivePropertyModel) getMultiple(query string, args ...any) ([]DiveProperty, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []DiveProperty
	for rows.Next() {
		var record DiveProperty
		err := rows.Scan(
			&record.ID,
			&record.Sort,
			&record.IsDefault,
			&record.Name,
			&record.Description,
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

	return records, nil
}

func (m *DivePropertyModel) List() ([]DiveProperty, error) {
	// If the list of equipment has already been populated, then use it.
	if len(divePropertyList) != 0 {
		return divePropertyList, nil
	}

	records, err := m.getMultiple(divePropertySelectQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to list all dive property items: %w", err)
	}

	// Cache the response for faster future calls.
	divePropertyList = records

	return records, nil
}

// Gets a slice of all the dive properties linked to a dive with the given ID.
func (m DivePropertyModel) GetAllForDive(diveID int) ([]DiveProperty, error) {
	var divePropertySelectQuery string = `
         select dp.id, dp.sort, dp.is_default, dp.name, dp.description
           from dive_dive_properties dd
     inner join dive_properties dp on dd.property_id = dp.id
          where dd.dive_id = $1
       order by dp.name
    `

	records, err := m.getMultiple(divePropertySelectQuery, diveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get properties for dive with id %d: %w", diveID, err)
	}

	return records, nil
}

// As property items are cached locally, it's likely more efficient to fetch
// the list ourselves and search for the given ID manually than to query the
// database directly.
func (m DivePropertyModel) Exists(id int) (bool, error) {
	items, err := m.List()
	if err != nil {
		msg := "failed to check if property with id %d exists: %w"
		return false, fmt.Errorf(msg, id, err)
	}

	for _, item := range items {
		if item.ID == id {
			return true, nil
		}
	}

	return false, nil
}

// AllExist checks that a record exists in the database for every ID in the
// given slice of `ids`.
func (m DivePropertyModel) AllExist(ids []int) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	stmt := `
        select count(id) = $1 as all_exist
          from dive_properties
         where id = any($2)
    `

	var allExist bool
	err := m.DB.QueryRowContext(ctx, stmt, len(ids), pq.Array(ids)).Scan(&allExist)
	if err != nil {
		msg := "failed to scan result of all ids (%v) exist check in dive_properties: %w"
		return false, fmt.Errorf(msg, ids, err)
	}

	return allExist, nil
}
