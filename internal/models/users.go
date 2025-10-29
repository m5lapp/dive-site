package models

import (
	"context"
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
	LastLogIn              time.Time
	DivingSince            time.Time
	DiveNumberOffset       int
	DivesLogged            int
	TotalDives             int
	MaxDiveNumber          int
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

	Update(
		userID int,
		name, email string,
		divingSince time.Time,
		diveNumberOffset, defaultDivingCountryID int,
		defaultDivingTZ TimeZone,
		darkMode bool,
	) error

	UpdatePassword(userID int, currentPassword, newPassword string) error
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var id int
	var hashedPassword []byte

	stmt := `select id, hashed_password from users
              where email = $1 and suspended = false and deleted = false`
	err := m.DB.QueryRowContext(ctx, stmt, email).Scan(&id, &hashedPassword)
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

	// Update the user's last log in date, we don't care if this errors though.
	stmt = `update users set last_log_in = now() where email = $1`
	_, _ = m.DB.ExecContext(ctx, stmt, email)

	return id, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	return idExistsInTable(m.DB, id, "users", "id")
}

func (m *UserModel) GetByID(id int) (User, error) {
	var user User

	stmt := `
        with user_dives as (
          select count(dv.id) dives_logged,
                 coalesce(max(dv.number), 0) max_dive_number
            from dives dv
           where owner_id = $1
        )
        select us.id, us.created_at, us.updated_at, us.name, us.friendly_name,
               us.email, us.suspended, us.deleted, us.last_log_in, us.dark_mode,
               us.diving_since, us.dive_number_offset,
               ud.dives_logged,
               ud.dives_logged + us.dive_number_offset total_dives,
               ud.max_dive_number,
               us.default_diving_country_id, us.default_diving_tz
          from users us
    cross join user_dives ud
         where us.id = $1
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
		&user.LastLogIn,
		&user.DarkMode,
		&user.DivingSince,
		&user.DiveNumberOffset,
		&user.DivesLogged,
		&user.MaxDiveNumber,
		&user.TotalDives,
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

func (m UserModel) UpdatePassword(userID int, currentPassword, newPassword string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var currentPasswordHashed []byte

	stmt := `select hashed_password from users
              where id = $1 and suspended = false and deleted = false`
	err := m.DB.QueryRowContext(ctx, stmt, userID).Scan(&currentPasswordHashed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidCredentials
		} else {
			return err
		}
	}

	err = bcrypt.CompareHashAndPassword(currentPasswordHashed, []byte(currentPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		} else {
			return err
		}
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	stmt = `
        update users
           set version = version + 1, updated_at = now(), hashed_password = $2
         where id = $1
    `

	result, err := m.DB.ExecContext(ctx, stmt, userID, newPasswordHash)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		if rowsAffected == 0 {
			return ErrNoRecord
		} else if rowsAffected > 1 {
			return &ErrUnexpectedRowsAffected{rowsExpected: 1, rowsAffected: int(rowsAffected)}
		}

		return err
	}

	return nil
}

func (m UserModel) Update(
	userID int,
	name, email string,
	divingSince time.Time,
	diveNumberOffset, defaultDivingCountryID int,
	defaultDivingTZ TimeZone,
	darkMode bool,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stmt := `
        update users
           set version = version + 1, updated_at = now(), name = $2,
               friendly_name = $2, email = $3, dark_mode = $4,
               diving_since = $5, dive_number_offset = $6,
               default_diving_country_id = $7, default_diving_tz = $8
         where id = $1
    `

	result, err := m.DB.ExecContext(
		ctx,
		stmt,
		userID,
		name,
		email,
		darkMode,
		divingSince,
		diveNumberOffset,
		defaultDivingCountryID,
		defaultDivingTZ,
	)
	if err != nil {
		return err
	}

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		if rowsAffected == 0 {
			return ErrNoRecord
		} else if rowsAffected > 1 {
			return &ErrUnexpectedRowsAffected{rowsExpected: 1, rowsAffected: int(rowsAffected)}
		}

		return err
	}

	return nil
}
