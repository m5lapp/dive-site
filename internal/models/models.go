package models

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

// The RowScanner interface allows both sql.Rows.Scan() and sql.Row.Scan() to be
// passed to a function for populating a struct from a database row.
type RowScanner interface {
	Scan(dest ...any) error
}

// TimeZone embeds a time.Location struct to represent a time zone location that
// can be saved and fetched from a SQL database.
type TimeZone struct {
	time.Location
}

func NewTimeZone(location string) (TimeZone, error) {
	l, err := time.LoadLocation(location)
	if err != nil {
		return TimeZone{}, fmt.Errorf("could not create TimeZone instance: %w", err)
	}

	return TimeZone{Location: *l}, nil
}

// String implements the fmt.Stringer interface.
func (tz TimeZone) String() string {
	return fmt.Sprint(&tz.Location)
}

// Scan implements the database/sql.Scanner interface. It takes a values from
// the database (hopefully a []byte or a string that represents a time.Location)
// and attempts to store it into the TimeZone struct.
func (tz *TimeZone) Scan(value any) error {
	if value == nil {
		return nil
	}

	var s string

	switch value.(type) {
	case []byte, string:
		s = value.(string)
	default:
		return fmt.Errorf("value provided to TimeZone.Scan must be a string or []byte")
	}

	l, err := time.LoadLocation(s)
	if err != nil {
		return err
	}

	tz.Location = *l
	return nil
}

// Value implements the database/sql/driver.Valuer interface. It returns a
// string representation of a time.Location from the TimeZone struct, suitable
// for storing in a SQL database.
func (tz TimeZone) Value() (driver.Value, error) {
	strValue := fmt.Sprint(&tz.Location)
	return []byte(strValue), nil
}

// idExistsInTable is a helper function that is intended to be used by any Model
// struct to efficiently check if a given integer ID exists in a given table
// using the given ID column name.
func idExistsInTable(db *sql.DB, id int, tableName, idColumn string) (bool, error) {
	stmt := fmt.Sprintf("select exists(select true from %s where %s = $1)", tableName, idColumn)

	var exists bool
	err := db.QueryRow(stmt, id).Scan(&exists)
	if err != nil {
		err = fmt.Errorf(
			"failed to check if id %d exists in %s.%s: %w",
			id,
			tableName,
			idColumn,
			err,
		)
	}

	return exists, err
}

// bindVarList returns a string of `count` comma-separated bind variable
// placeholders starting at `start` suitable for use in a PostgreSQL query. It
// returns the empty string if count is less than 1.
func bindVarList(start, count int) (string, error) {
	if count < 1 {
		return "", nil
	}

	if start < 1 {
		msg := "cannot use %d as start index for bind variables, must start from at least 1"
		return "", fmt.Errorf(msg, start)
	}

	var s strings.Builder
	fmt.Fprintf(&s, "$%d", start)
	for i := start + 1; i < start+count; i++ {
		fmt.Fprintf(&s, ", $%d", i)
	}

	return s.String(), nil
}

func buildManyToManyInsert(
	intermediateTable, parentCol, childCol string,
	rowCount int,
) (string, error) {
	if rowCount < 1 {
		msg := "cannot create insert query for %s with %d rows, needs at least 1"
		return "", fmt.Errorf(msg, intermediateTable, rowCount)
	}

	var stmt strings.Builder
	_, err := fmt.Fprintf(
		&stmt,
		"insert into %s (%s, %s) values",
		intermediateTable,
		parentCol,
		childCol,
	)

	if err != nil {
		return "", err
	}

	for i := range rowCount {
		fmt.Fprintf(&stmt, " ($1, $%d),", i+2)
	}

	// Strip off the trailing comma.
	return strings.TrimRight(stmt.String(), ","), nil
}

// buildManyToManyUpsert returns a SQL query that will insert or update all the
// rows in `intermediateTable` for a given ID in `parentCol` using placeholder
// $1 with the ID values in `childCol` which is a placeholder array identified
// by $2. The query does not work if $2 is an empty array so this scenario
// should be guarded against.
//
// When the query is run, any values in $2 that do not already exist in the
// intermediate table will be added to it, any that are in the intermediate
// table but missing from $2 will be deleted from the table and any that exist
// in both will be left unaffected.
func buildManyToManyUpsert(intermediateTable, parentCol, childCol string) string {
	query := `
      -- Create a temporary table that maps the ID in the parent column to each
      -- ID in the child column that needs to be upserted.
      with new_vals as (
        select $1::numeric as %[2]s,
               unnest($2::numeric[]) as %[3]s
      ),
      -- This CTE deletes any values from the intermediate table (it) that do
      -- not exist in the new_vals temporary table.
      delete_vals_to_remove as (
        delete from %[1]s it
         using new_vals nv
         where it.%[2]s = $1::numeric
           and it.%[3]s not in (select %[3]s from new_vals)
      )
      -- Finally, insert the values from the new_vals table into the
      -- intermediate table. Any conflicts can be safely ignored as these are
      -- just existing rows that need to be kept.
      insert into %[1]s (%[2]s, %[3]s)
      select %[2]s, %[3]s from new_vals
 on conflict do nothing
    `

	return fmt.Sprintf(query, intermediateTable, parentCol, childCol)
}

func sliceToAnySlice[T any](values []T) []any {
	var result []any = make([]any, len(values))

	if values == nil {
		return result
	}

	for i, value := range values {
		result[i] = value
	}

	return result
}

// sqlExecer is implemented by all three of sql.Conn, sql.DB and sql.Tx.
type sqlExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type sqlID interface {
	int | int8 | int32 | int64 | uint | uint8 | uint32 | uint64
}

func insertManyToManyIDs[T sqlID](
	ctx context.Context,
	db sqlExecer,
	tableName, parentField, childField string,
	parentID T,
	childIDs []T,
) error {
	if len(childIDs) == 0 {
		return nil
	}

	stmt, err := buildManyToManyInsert(tableName, parentField, childField, len(childIDs))
	if err != nil {
		errMsg := "failed to build insert query for %s: %w"
		return fmt.Errorf(errMsg, tableName, err)
	}

	// In order to unpack the variadic parameters into the query function,
	// we need to first combine them all into a single []any slice.
	args := make([]any, 0, len(childIDs)+1)
	args = append(args, parentID)
	args = append(args, sliceToAnySlice(childIDs)...)

	_, err = db.ExecContext(ctx, stmt, args...)
	if err != nil {
		errMsg := "failed to insert many-to-many ids (%v, %v) for %s: %w"
		return fmt.Errorf(errMsg, parentID, childIDs, tableName, err)
	}

	return nil
}

func upsertManyToManyIDs[T sqlID](
	ctx context.Context,
	db sqlExecer,
	tableName, parentCol, childCol string,
	parentID T,
	childIDs []T,
) error {
	var err error

	if len(childIDs) == 0 {
		query := fmt.Sprintf("delete from %s where %s = $1", tableName, parentCol)
		_, err = db.ExecContext(ctx, query, parentID)
	} else {
		query := buildManyToManyUpsert(tableName, parentCol, childCol)
		_, err = db.ExecContext(ctx, query, parentID, pq.Array(childIDs))
	}

	if err != nil {
		errMsg := "failed to upsert many-to-many ids (%v, %v) for %s: %w"
		return fmt.Errorf(errMsg, parentID, childIDs, tableName, err)
	}

	return nil
}
