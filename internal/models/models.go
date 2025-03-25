package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// The RowScanner interface allows both sql.Rows.Scan() and sql.Row.Scan() to be
// passed to a function for populating a struct from a database row.
type RowScanner interface {
	Scan(dest ...any) error
}

// TimeZone embedds a time.Location struct to represent a time zone location
// that can be saved and fetched from a SQL database.
type TimeZone struct {
	time.Location
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
// a string representation of a time.Location from the TimeZone struct, suitable
// for storing in a SQL database.
func (tz TimeZone) Value() (driver.Value, error) {
	strValue := fmt.Sprint(&tz.Location)
	return []byte(strValue), nil
}
