package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type EquipmentModel struct {
	DB *sql.DB
}

type EquipmentModelInterface interface {
	AllExist(ids []int) (bool, error)
	Exists(id int) (bool, error)
	List() ([]Equipment, error)
}

type Equipment struct {
	ID          int
	Sort        int
	IsDefault   bool
	Name        string
	Description string
}

func (e Equipment) String() string {
	return fmt.Sprintf("%s - %s", e.Name, e.Description)
}

// equipmentList stores a static, cached slice of Equipment data so that
// successive requests can bypass the database call.
var equipmentList []Equipment

var equipmentSelectQuery string = `
     select eq.id, eq.sort, eq.is_default, eq.name, eq.description
       from equipment eq
   order by eq.name
`

func (m *EquipmentModel) getMultiple(query string, args ...any) ([]Equipment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Equipment
	for rows.Next() {
		var record Equipment
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

func (m *EquipmentModel) List() ([]Equipment, error) {
	// If the list of equipment has already been populated, then use it.
	if len(equipmentList) != 0 {
		return equipmentList, nil
	}

	records, err := m.getMultiple(equipmentSelectQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to list all equipment items: %w", err)
	}

	// Cache the response for faster future calls.
	equipmentList = records

	return records, nil
}

// Gets a slice of all the equipment items linked to a dive with the given ID.
func (m EquipmentModel) GetAllForDive(diveID int) ([]Equipment, error) {
	var diveEquipmentSelectQuery string = `
         select eq.id, eq.sort, eq.is_default, eq.name, eq.description
           from dive_equipment de
     inner join equipment eq on de.equipment_id = eq.id
          where de.dive_id = $1
       order by eq.name
    `

	records, err := m.getMultiple(diveEquipmentSelectQuery, diveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment for dive with id %d: %w", diveID, err)
	}

	return records, nil
}

// As equipment items are cached locally, it's likely more efficient to fetch
// the list ourselves and search for the given ID manually than to query the
// database directly.
func (m EquipmentModel) Exists(id int) (bool, error) {
	items, err := m.List()
	if err != nil {
		msg := "failed to check if equipment with id %d exists: %w"
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
func (m EquipmentModel) AllExist(ids []int) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	stmt := `
        select count(id) = $1 as all_exist
          from equipment
         where id = any($2)
    `

	var allExist bool
	err := m.DB.QueryRowContext(ctx, stmt, len(ids), pq.Array(ids)).Scan(&allExist)
	if err != nil {
		msg := "failed to scan result of all ids (%v) exist check in equipment: %w"
		return false, fmt.Errorf(msg, ids, err)
	}

	return allExist, nil
}
