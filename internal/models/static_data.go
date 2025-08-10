package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/lib/pq"
)

type staticDataInterface interface {
	getValuesFromDBRow(
		rs RowScanner,
	) (id, sort int, isDefault bool, name, description string, err error)

	id() int

	setValues(id, sort int, isDefault bool, name, description string)

	tableName() string
}

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

func (sd staticDataItem) id() int {
	return sd.ID
}

func (sd staticDataItem) getValuesFromDBRow(
	rs RowScanner,
) (id, sort int, isDefault bool, name, description string, err error) {
	item := &staticDataItem{}

	err = rs.Scan(&item.ID, &item.Sort, &item.IsDefault, &item.Name, &item.Description)
	if err != nil {
		return 0, 0, false, "", "", err
	}

	return item.ID, item.Sort, item.IsDefault, item.Name, item.Description, err
}

func (sd staticDataItem) setValues(id, sort int, isDefault bool, name, description string) {
	sd.ID = id
	sd.Sort = sort
	sd.IsDefault = isDefault
	sd.Name = name
	sd.Description = description

	// This just gets rid of the annoying "unused write to field" warnings.
	_, _, _, _, _ = sd.ID, sd.Sort, sd.IsDefault, sd.Name, sd.Description
}

type nullableStaticDataItem struct {
	ID          *int
	Sort        *int
	IsDefault   *bool
	Name        *string
	Description *string
}

func nullableStaticDataItemToStruct[T staticDataInterface](ns nullableStaticDataItem) *T {
	if ns.ID == nil {
		return nil
	}

	// Check none of the values are null, use the zero value if they are.
	if ns.Sort == nil {
		*ns.Sort = 0
	}
	if ns.IsDefault == nil {
		*ns.IsDefault = false
	}
	if ns.Name == nil {
		*ns.Name = ""
	}
	if ns.Description == nil {
		*ns.Description = ""
	}

	var item T
	item.setValues(*ns.ID, *ns.Sort, *ns.IsDefault, *ns.Name, *ns.Description)

	return &item
}

type StaticDataService[T staticDataInterface] struct {
	DB *sql.DB
}

func NewStaticDataService[T staticDataInterface](db *sql.DB) *StaticDataService[T] {
	return &StaticDataService[T]{DB: db}
}

// staticData is a map which stores a cached slice of each of the static data
// types using the static data's table as the key. This allows successive
// requests to bypass the database call.
var staticData map[string][]staticDataInterface = make(map[string][]staticDataInterface)

func getCachedStaticData[T staticDataInterface]() ([]T, error) {
	// We need this to be able to get the tableName value for the type T.
	var tempT T
	tableName := tempT.tableName()

	data, ok := staticData[tableName]
	if !ok {
		return nil, fmt.Errorf("could not find key %s in map", tableName)
	}

	if len(data) == 0 {
		return []T{}, nil
	}

	// Do a quick sanity check to ensure that the first element of the list is
	// of type T. As this is an internal cache, it should be, but it's good to
	// check.
	expectedType := reflect.TypeOf((*T)(nil)).Elem()
	firstElementType := reflect.TypeOf(data[0])
	if firstElementType != expectedType {
		return nil, fmt.Errorf(
			"unexpected type mismatch, expected %v, but got %v for key %s",
			expectedType,
			firstElementType,
			tableName,
		)
	}

	// If everything seems OK, then we need to cast each item in data to the
	//concrete type T before we can return it.
	typedData := make([]T, len(data))
	for i, v := range data {
		typedData[i] = v.(T)
	}
	return typedData, nil
}

var staticDataItemSelectQuery string = `
    select si.id, si.sort, si.is_default, si.name, si.description
      from %s si
  order by %s
`

// As staticDataItems are cached locally, it's likely more efficient to fetch
// the list ourselves and search for the given ID manually than to query the
// database directly.
func (m *StaticDataService[T]) Exists(id int) (bool, error) {
	items, err := m.List(false)
	if err != nil {
		msg := "failed to check if %T with id %d exists: %w"
		return false, fmt.Errorf(msg, *new(T), id, err)
	}

	for _, item := range items {
		if item.id() == id {
			return true, nil
		}
	}

	return false, nil
}

func (m *StaticDataService[T]) AllExist(ids []int) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}

	// We need this to be able to get the tableName value for the type T.
	var tempT T
	tableName := tempT.tableName()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	stmt := `
        select count(id) = $1 as all_exist
          from ` + tableName + `
         where id = any($2)
    `

	var allExist bool
	err := m.DB.QueryRowContext(ctx, stmt, len(ids), pq.Array(ids)).Scan(&allExist)
	if err != nil {
		msg := "failed to scan result of all ids (%v) exist check in %s: %w"
		return false, fmt.Errorf(msg, ids, tableName, err)
	}

	return allExist, nil
}

func (m *StaticDataService[T]) GetOneByID(id int) (T, error) {
	var emptyT T

	items, err := m.List(false)
	if err != nil {
		msg := "failed to fetch %T with id %d: %w"
		return emptyT, fmt.Errorf(msg, emptyT, id, err)
	}

	for _, item := range items {
		if item.id() == id {
			return item, nil
		}
	}

	return emptyT, ErrNoRecord
}

func (m *StaticDataService[T]) List(sortByName bool) ([]T, error) {
	// We need this to be able to get the tableName value for the type T.
	var tempT T
	tableName := tempT.tableName()

	// If the list of static data items has already been populated, then use it.
	// Otherwise, if there was an error, we don't really care and will try to
	// load the values from the database.
	data, err := getCachedStaticData[T]()
	if err == nil && len(data) >= 0 {
		return data, nil
	}

	sortColumn := "sort"
	if sortByName {
		sortColumn = "name"
	}

	stmt := fmt.Sprintf(staticDataItemSelectQuery, tableName, sortColumn)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch static data from %s: %w", tableName, err)
	}
	defer rows.Close()

	records := []T{}
	for rows.Next() {
		var record T
		id, sort, isDefault, name, description, err := record.getValuesFromDBRow(rows)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to scan static data row from table %s: %w",
				tableName,
				err,
			)
		}

		// We use reflect here to check for and set the various fields of the
		// static data item. This isn't ideal, but was the only way to get this
		// working when using a value receiver.
		recordPtr := &record
		v := reflect.ValueOf(recordPtr).Elem()
		if v.Kind() != reflect.Struct {
			return nil, fmt.Errorf("expected struct, but got a %d (%v)", v.Kind(), record)
		}

		idField := v.FieldByName("ID")
		sortField := v.FieldByName("Sort")
		isDefaultField := v.FieldByName("IsDefault")
		nameField := v.FieldByName("Name")
		descriptionField := v.FieldByName("Description")

		if !idField.IsValid() || !sortField.IsValid() || !isDefaultField.IsValid() ||
			!nameField.IsValid() || !descriptionField.IsValid() {
			msg := "%T struct does not have requisite fields ID, Sort, IsDefault, Name, Description"
			return nil, fmt.Errorf(msg, record)
		}

		if !idField.CanSet() || !sortField.CanSet() || !isDefaultField.CanSet() ||
			!nameField.CanSet() || !descriptionField.CanSet() {
			msg := "field(s) in struct %T are not settable"
			return nil, fmt.Errorf(msg, record)
		}

		if idField.Kind() != reflect.Int ||
			sortField.Kind() != reflect.Int ||
			isDefaultField.Kind() != reflect.Bool ||
			nameField.Kind() != reflect.String ||
			descriptionField.Kind() != reflect.String {
			msg := "field(s) in struct %T are not of required types"
			return nil, fmt.Errorf(msg, record)
		}

		idField.SetInt(int64(id))
		sortField.SetInt(int64(sort))
		isDefaultField.SetBool(isDefault)
		nameField.SetString(name)
		descriptionField.SetString(description)

		records = append(records, record)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// Cache the response for faster future calls. We need to convert the []T
	// to a []staticDataInterface first though, even though T implements the
	// interface.
	interfaceData := make([]staticDataInterface, len(records))
	for i, v := range records {
		interfaceData[i] = v
	}
	staticData[tableName] = interfaceData

	return records, nil
}

// Here we create our static data structs for the different types of data and
// embed the staticDataItem base struct into them.

// Current.
type CurrentModelInterface interface {
	Exists(id int) (bool, error)
	List(sortByName bool) ([]Current, error)
}

type Current struct {
	staticDataItem
}

func (_ Current) tableName() string {
	return "currents"
}

// DiveProperty.
type DivePropertyModelInterface interface {
	AllExist(ids []int) (bool, error)
	Exists(id int) (bool, error)
	List(sortByName bool) ([]DiveProperty, error)
}

type DiveProperty struct {
	staticDataItem
}

func (_ DiveProperty) tableName() string {
	return "dive_properties"
}

// Entry point.
type EntryPointModelInterface interface {
	Exists(id int) (bool, error)
	List(sortByName bool) ([]EntryPoint, error)
}

type EntryPoint struct {
	staticDataItem
}

func (_ EntryPoint) tableName() string {
	return "entry_points"
}

// Equipment.
type EquipmentModelInterface interface {
	AllExist(ids []int) (bool, error)
	Exists(id int) (bool, error)
	List(sortByName bool) ([]Equipment, error)
}

type Equipment struct {
	staticDataItem
}

func (_ Equipment) tableName() string {
	return "equipment"
}

// Gas mix.
type GasMixModelInterface interface {
	Exists(id int) (bool, error)
	GetOneByID(id int) (GasMix, error)
	List(sortByName bool) ([]GasMix, error)
}

type GasMix struct {
	staticDataItem
}

func (_ GasMix) tableName() string {
	return "gas_mixes"
}

// Tank configuration.
type TankConfigurationModelInterface interface {
	Exists(id int) (bool, error)
	List(sortByName bool) ([]TankConfiguration, error)
}

type TankConfiguration struct {
	staticDataItem
}

func (_ TankConfiguration) tableName() string {
	return "tank_configurations"
}

// Tank material.
type TankMaterialModelInterface interface {
	Exists(id int) (bool, error)
	List(sortByName bool) ([]TankMaterial, error)
}

type TankMaterial struct {
	staticDataItem
}

func (_ TankMaterial) tableName() string {
	return "tank_materials"
}

// Waves.
type WavesModelInterface interface {
	Exists(id int) (bool, error)
	List(sortByName bool) ([]Waves, error)
}

type Waves struct {
	staticDataItem
}

func (_ Waves) tableName() string {
	return "waves"
}
