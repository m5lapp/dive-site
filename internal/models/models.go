package models

// The RowScanner interface allows both sql.Rows.Scan() and sql.Row.Scan() to be
// passed to a function for populating a struct from a database row.
type RowScanner interface {
	Scan(dest ...any) error
}
