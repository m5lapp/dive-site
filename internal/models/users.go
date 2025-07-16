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
	Updated                time.Time
	Name                   string
	FriendlyName           string
	Email                  string
	HashedPassword         []byte
	Suspended              bool
	Deleted                bool
	DivingSince            time.Time
	DiveNumberOffset       int
	DefaultDivingCountryID int
	DefaultDivingTZ        TimeZone
	DarkMode               bool
}

var AnonymousUser = &User{}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
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
	GetByID(id int) (User, error)
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

	stmt := `select id, hashed_password from users
              where email = $1 and suspended = false and deleted = false`
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
	return idExistsInTable(m.DB, id, "users", "id")
}

func (m *UserModel) GetByID(id int) (User, error) {
	var user User

	stmt := `
        select id, created_at, updated_at, name, friendly_name, email,
               suspended, deleted, dark_mode, diving_since, dive_number_offset,
               default_diving_country_id, default_diving_tz
          from users where id = $1
    `

	err := m.DB.QueryRow(stmt, id).Scan(
		&user.ID,
		&user.Created,
		&user.Updated,
		&user.Name,
		&user.FriendlyName,
		&user.Email,
		&user.Suspended,
		&user.Deleted,
		&user.DarkMode,
		&user.DivingSince,
		&user.DiveNumberOffset,
		&user.DefaultDivingCountryID,
		&user.DefaultDivingTZ,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNoRecord
		} else {
			return User{}, err
		}
	}

	return user, nil
}
