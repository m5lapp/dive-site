package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type staticDataItem struct {
	ID          int
	Sort        int
	IsDefault   bool
	Name        string
	Description string
}

func (sd staticDataItem) String() string {
	return fmt.Sprintf("%s - %s", sd.Name, sd.Description)
}

type staticDataItemTable string

var staticDataItemsCache map[staticDataItemTable][]staticDataItem = make(
	map[staticDataItemTable][]staticDataItem,
)

var staticDataItemSelectQuery string = `
    select si.id, si.sort, si.is_default, si.name, si.description
      from %s si
  order by si.%s
`

func listStaticDataItems(
	db *sql.DB,
	table staticDataItemTable,
	sortByName bool,
) ([]staticDataItem, error) {
	// If the list of static data items has already been populated, then use it.
	// Otherwise, we don't really care and will try to load the values from the
	// database.
	data, ok := staticDataItemsCache[table]
	if ok && len(data) >= 0 {
		return data, nil
	}

	sortColumn := "sort"
	if sortByName {
		sortColumn = "name"
	}
	stmt := fmt.Sprintf(staticDataItemSelectQuery, table, sortColumn)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []staticDataItem
	for rows.Next() {
		var record staticDataItem
		err = rows.Scan(
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

	// Cache the response for faster future calls.
	staticDataItemsCache[table] = records

	return records, nil
}

type nullableStaticDataItem struct {
	ID          *int
	Sort        *int
	IsDefault   *bool
	Name        *string
	Description *string
}

func (ns nullableStaticDataItem) ToStruct() *staticDataItem {
	if ns.ID == nil {
		return nil
	}

	return &staticDataItem{
		ID:          *ns.ID,
		Sort:        *ns.Sort,
		IsDefault:   *ns.IsDefault,
		Name:        *ns.Name,
		Description: *ns.Description,
	}
}

func (ns nullableStaticDataItem) ToCurrent() *Current {
	sdi := ns.ToStruct()

	if sdi == nil {
		return nil
	}

	return &Current{
		staticDataItem: *sdi,
	}
}

func (ns nullableStaticDataItem) ToWaves() *Waves {
	sdi := ns.ToStruct()

	if sdi == nil {
		return nil
	}

	return &Waves{
		staticDataItem: *sdi,
	}
}

// Current.

type CurrentModelInterface interface {
	Exists(id int) (bool, error)
	List(sortByName bool) ([]Current, error)
}

type Current struct {
	staticDataItem
}

type CurrentModel struct {
	DB *sql.DB
}

const currentTable staticDataItemTable = "currents"

func (m *CurrentModel) Exists(id int) (bool, error) {
	return idExistsInTable(m.DB, id, string(currentTable), "id")
}

func (m *CurrentModel) List(sortByName bool) ([]Current, error) {
	staticDataItems, err := listStaticDataItems(m.DB, currentTable, sortByName)

	if err != nil {
		return nil, fmt.Errorf("failed to list items from table %s: %w", currentTable, err)
	}

	var items []Current
	for _, item := range staticDataItems {
		items = append(items, Current{staticDataItem: item})
	}

	return items, nil
}

// Entry point.

type EntryPointModelInterface interface {
	Exists(id int) (bool, error)
	List(sortByName bool) ([]EntryPoint, error)
}

type EntryPoint struct {
	staticDataItem
}

type EntryPointModel struct {
	DB *sql.DB
}

const entrypointTable staticDataItemTable = "entry_points"

func (m *EntryPointModel) Exists(id int) (bool, error) {
	return idExistsInTable(m.DB, id, string(entrypointTable), "id")
}

func (m *EntryPointModel) List(sortByName bool) ([]EntryPoint, error) {
	staticDataItems, err := listStaticDataItems(m.DB, entrypointTable, sortByName)

	if err != nil {
		return nil, fmt.Errorf("failed to list items from table %s: %w", entrypointTable, err)
	}

	var items []EntryPoint
	for _, item := range staticDataItems {
		items = append(items, EntryPoint{staticDataItem: item})
	}

	return items, nil
}

// Gas Mix.

type GasMixModelInterface interface {
	Exists(id int) (bool, error)
	GetOneByID(id int) (GasMix, error)
	List(sortByName bool) ([]GasMix, error)
}

type GasMix struct {
	staticDataItem
}

type GasMixModel struct {
	DB *sql.DB
}

const gasMixTable staticDataItemTable = "gas_mixes"

func (m *GasMixModel) Exists(id int) (bool, error) {
	return idExistsInTable(m.DB, id, string(entrypointTable), "id")
}

func (m *GasMixModel) GetOneByID(id int) (GasMix, error) {
	// As staticDataItems are cached, we may as well get the cached list and
	// search for the gas mix with the ID there rather than making a DB call.
	items, err := m.List(false)
	if err != nil {
		msg := "failed to fetch gas mix with id %d: %w"
		return GasMix{}, fmt.Errorf(msg, id, err)
	}

	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}

	return GasMix{}, ErrNoRecord
}

func (m *GasMixModel) List(sortByName bool) ([]GasMix, error) {
	staticDataItems, err := listStaticDataItems(m.DB, gasMixTable, sortByName)

	if err != nil {
		return nil, fmt.Errorf("failed to list items from table %s: %w", gasMixTable, err)
	}

	var items []GasMix
	for _, item := range staticDataItems {
		items = append(items, GasMix{staticDataItem: item})
	}

	return items, nil
}

// Tank Configuration,

type TankConfigurationModelInterface interface {
	Exists(id int) (bool, error)
	List(sortByName bool) ([]TankConfiguration, error)
}

type TankConfiguration struct {
	staticDataItem
	TankCount int
}

type TankConfigurationModel struct {
	DB *sql.DB
}

const tankConfigurationTable staticDataItemTable = "tank_configurations"

func (m *TankConfigurationModel) Exists(id int) (bool, error) {
	return idExistsInTable(m.DB, id, string(tankConfigurationTable), "id")
}

func (m *TankConfigurationModel) List(sortByName bool) ([]TankConfiguration, error) {
	sortColumn := "sort"
	if sortByName {
		sortColumn = "name"
	}
	stmt := `
    select si.id, si.sort, si.is_default, si.name, si.description, si.tank_count
      from %s si
  order by si.%s
`
	stmt = fmt.Sprintf(stmt, tankConfigurationTable, sortColumn)

	errMsg := "failed to list items from table %s: %w"
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf(errMsg, tankConfigurationTable, err)
	}
	defer rows.Close()

	var records []TankConfiguration
	for rows.Next() {
		var record TankConfiguration
		err = rows.Scan(
			&record.ID,
			&record.Sort,
			&record.IsDefault,
			&record.Name,
			&record.Description,
			&record.TankCount,
		)
		if err != nil {
			return nil, fmt.Errorf(errMsg, tankConfigurationTable, err)
		}
		records = append(records, record)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf(errMsg, tankConfigurationTable, err)
	}

	return records, nil
}

// Tank Material.

type TankMaterialModelInterface interface {
	Exists(id int) (bool, error)
	List(sortByName bool) ([]TankMaterial, error)
}

type TankMaterial struct {
	staticDataItem
}

type TankMaterialModel struct {
	DB *sql.DB
}

const tankMaterialTable staticDataItemTable = "tank_materials"

func (m *TankMaterialModel) Exists(id int) (bool, error) {
	return idExistsInTable(m.DB, id, string(tankMaterialTable), "id")
}

func (m *TankMaterialModel) List(sortByName bool) ([]TankMaterial, error) {
	staticDataItems, err := listStaticDataItems(m.DB, tankMaterialTable, sortByName)

	if err != nil {
		return nil, fmt.Errorf("failed to list items from table %s: %w", tankMaterialTable, err)
	}

	var items []TankMaterial
	for _, item := range staticDataItems {
		items = append(items, TankMaterial{staticDataItem: item})
	}

	return items, nil
}

// Waves.

type WavesModelInterface interface {
	Exists(id int) (bool, error)
	List(sortByName bool) ([]Waves, error)
}

type Waves struct {
	staticDataItem
}

type WavesModel struct {
	DB *sql.DB
}

const wavesTable staticDataItemTable = "waves"

func (m *WavesModel) Exists(id int) (bool, error) {
	return idExistsInTable(m.DB, id, string(wavesTable), "id")
}

func (m *WavesModel) List(sortByName bool) ([]Waves, error) {
	staticDataItems, err := listStaticDataItems(m.DB, wavesTable, sortByName)

	if err != nil {
		return nil, fmt.Errorf("failed to list items from table %s: %w", wavesTable, err)
	}

	var items []Waves
	for _, item := range staticDataItems {
		items = append(items, Waves{staticDataItem: item})
	}

	return items, nil
}
