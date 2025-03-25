package models

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                     int
	Created                time.Time
	Update                 time.Time
	Name                   string
	FriendlyName           string
	Email                  string
	HashedPassword         []byte
	DivingSince            time.Time
	DiveNumberOffset       int
	DefaultDivingCountryID int
	DefaultDivingTZ        TimeZone
	DarkMode               bool
}

type UserModelInterface interface {
	Insert(
		name, email, password string,
		divingSince time.Time,
		diveNumberOffset, defaultDivingCountryID int,
		defaultDivingTZ TimeZone,
		darkMode bool,
	) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(
	name, email, password string,
	divingSince time.Time,
	diveNumberOffset, defaultDivingCountryID int,
	defaultDivingTZ TimeZone,
	darkMode bool,
) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `
        insert into users (
             name, friendly_name, email, hashed_password, dark_mode,
             diving_since, dive_number_offset, default_diving_country_id,
             default_diving_tz
        ) values ($1, $1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err = m.DB.Exec(
		stmt,
		name,
		email,
		hashedPassword,
		darkMode,
		divingSince,
		diveNumberOffset,
		defaultDivingCountryID,
		defaultDivingTZ,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	stmt := `select id, hashed_password from users where email = $1`
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool

	stmt := `select exists(select true from users where id = $1)`
	err := m.DB.QueryRow(stmt, id).Scan(&exists)

	return exists, err
}
