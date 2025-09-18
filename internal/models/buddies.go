package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Buddy struct {
	ID              int
	Version         int
	Created         time.Time
	Updated         time.Time
	OwnerID         int
	Name            string
	Email           string
	PhoneNumber     string
	Agency          *Agency
	AgencyMemberNum string
	DivesWith       int
	FirstDiveWith   *time.Time
	LastDiveWith    *time.Time
	Notes           string
}

func (bu Buddy) String() string {
	var str strings.Builder
	str.WriteString(bu.Name)

	if bu.Agency != nil {
		str.WriteString(" (" + bu.Agency.Acronym)
		if bu.AgencyMemberNum != "" {
			str.WriteString(" #" + bu.AgencyMemberNum)
		}
		str.WriteString(")")
	}

	return str.String()
}

type nullableBuddy struct {
	ID              *int
	Version         *int
	Created         *time.Time
	Updated         *time.Time
	OwnerID         *int
	Name            *string
	Email           *string
	PhoneNumber     *string
	Agency          nullableAgency
	AgencyMemberNum *string
	DivesWith       *int
	FirstDiveWith   *time.Time
	LastDiveWith    *time.Time
	Notes           *string
}

func (nb nullableBuddy) ToStruct() *Buddy {
	if nb.ID == nil {
		return nil
	}

	return &Buddy{
		ID:              *nb.ID,
		Version:         *nb.Version,
		Created:         *nb.Created,
		Updated:         *nb.Updated,
		OwnerID:         *nb.OwnerID,
		Name:            *nb.Name,
		Email:           *nb.Email,
		PhoneNumber:     *nb.PhoneNumber,
		Agency:          nb.Agency.ToStruct(),
		AgencyMemberNum: *nb.AgencyMemberNum,
		DivesWith:       *nb.DivesWith,
		FirstDiveWith:   nb.FirstDiveWith,
		LastDiveWith:    nb.LastDiveWith,
		Notes:           *nb.Notes,
	}
}

type BuddyModelInterface interface {
	Exists(id int) (bool, error)
	Insert(
		ownerID int,
		name string,
		emailAddress string,
		phoneNumber string,
		agencyID *int,
		agencyMemberNum string,
		notes string,
	) (int, error)
	List(userID int, filters ListFilters) ([]Buddy, PageData, error)
	ListAll(userID int) ([]Buddy, error)
}

var buddySelectQuery string = `
      with buddy_dive_stats as (
        select dv.buddy_id buddy_id, count(dv.id) dives_with,
               min(dv.date_time_in) first_dive_with,
               max(dv.date_time_in) last_dive_with
          from dives dv
         where dv.owner_id = $1
      group by dv.buddy_id
           )
    select count(*) over(),
           bu.id, bu.version, bu.created_at, bu.updated_at, bu.owner_id,
           bu.name, bu.email, bu.phone_number,
           bu.agency_id, ag.common_name, ag.full_name, ag.acronym, ag.url,
           bu.agency_member_num,
           coalesce(ds.dives_with, 0), ds.first_dive_with, ds.last_dive_with,
           bu.notes
      from buddies bu
 left join agencies ag on bu.agency_id = ag.id
 left join buddy_dive_stats ds on bu.id = ds.buddy_id
     where bu.owner_id = $1
`

func buddyFromDBRow(rs RowScanner, totalRecords *int, bu *Buddy) error {
	ag := &nullableAgency{}

	err := rs.Scan(
		totalRecords,
		&bu.ID,
		&bu.Version,
		&bu.Created,
		&bu.Updated,
		&bu.OwnerID,
		&bu.Name,
		&bu.Email,
		&bu.PhoneNumber,
		&ag.ID,
		&ag.CommonName,
		&ag.FullName,
		&ag.Acronym,
		&ag.URL,
		&bu.AgencyMemberNum,
		&bu.DivesWith,
		&bu.FirstDiveWith,
		&bu.LastDiveWith,
		&bu.Notes,
	)

	if err != nil {
		return err
	}

	bu.Agency = ag.ToStruct()
	return nil
}

type BuddyModel struct {
	DB *sql.DB
}

func (m *BuddyModel) Exists(id int) (bool, error) {
	return idExistsInTable(m.DB, id, "buddies", "id")
}

func (m *BuddyModel) Insert(
	ownerID int,
	name string,
	emailAddress string,
	phoneNumber string,
	agencyID *int,
	agencyMemberNum string,
	notes string,
) (int, error) {
	stmt := `
        insert into buddies (
            owner_id, name, email, phone_number, agency_id, agency_member_num,
            notes
        ) values (
            $1, $2, $3, $4, $5, $6, $7
        )
        returning id
    `

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result := m.DB.QueryRowContext(
		ctx,
		stmt,
		ownerID,
		name,
		emailAddress,
		phoneNumber,
		agencyID,
		agencyMemberNum,
		notes,
	)

	var id int
	err := result.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *BuddyModel) List(userID int, filters ListFilters) ([]Buddy, PageData, error) {
	limit := filters.limit()
	offset := filters.offset()
	stmt := fmt.Sprintf("%s limit $2 offset $3", buddySelectQuery)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt, userID, limit, offset)
	if err != nil {
		return nil, PageData{}, err
	}
	defer rows.Close()

	var totalRecords int
	records := []Buddy{}
	for rows.Next() {
		var record Buddy
		err := buddyFromDBRow(rows, &totalRecords, &record)
		if err != nil {
			return nil, PageData{}, err
		}
		records = append(records, record)
	}

	err = rows.Err()
	if err != nil {
		return nil, PageData{}, err
	}

	paginationData := newPaginationData(
		totalRecords,
		filters.page,
		filters.pageSize,
	)

	return records, paginationData, nil
}

func (m *BuddyModel) ListAll(userID int) ([]Buddy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, buddySelectQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalRecords int
	var records []Buddy
	for rows.Next() {
		var record Buddy
		err := buddyFromDBRow(rows, &totalRecords, &record)
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

type BuddyRole struct {
	ID          int
	Name        string
	Description string
}

type nullableBuddyRole struct {
	ID          *int
	Name        *string
	Description *string
}

func (nb nullableBuddyRole) ToStruct() *BuddyRole {
	if nb.ID == nil {
		return nil
	}

	return &BuddyRole{
		ID:          *nb.ID,
		Name:        *nb.Name,
		Description: *nb.Description,
	}
}

type BuddyRoleModel struct {
	DB *sql.DB
}

type BuddyRoleModelInterface interface {
	Exists(id int) (bool, error)
	List() ([]BuddyRole, error)
}

// buddyRoleList stores a static, cached slice of BuddyRole data so that
// successive requests can bypass the database call.
var buddyRoleList []BuddyRole
var buddyRoleListQuery string = `
    select id, name, description
      from buddy_roles
  order by name
`

func (m *BuddyRoleModel) Exists(id int) (bool, error) {
	return idExistsInTable(m.DB, id, "buddy_roles", "id")
}

func (m *BuddyRoleModel) List() ([]BuddyRole, error) {
	// If the list of buddy roles has already been populated, then use it.
	if len(buddyRoleList) != 0 {
		return buddyRoleList, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, buddyRoleListQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []BuddyRole
	for rows.Next() {
		var record BuddyRole
		err := rows.Scan(&record.ID, &record.Name, &record.Description)
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
	buddyRoleList = records

	return buddyRoleList, nil
}
