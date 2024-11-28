package models

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func newTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open(
		"postgres",
		"postgres://divesite_integration_test:password@localhost:5432/divesite_integration_test?sslmode=disable",
	)
	if err != nil {
		t.Fatal(err)
	}

	script, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	// This function will automatically be called when the test finishes
	// running.
	t.Cleanup(func() {
		defer db.Close()

		script, err := os.ReadFile("./testdata/teardown.sql")
		if err != nil {
			t.Fatal(err)
		}

		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}
	})

	return db
}
