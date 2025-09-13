package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Certification struct {
	ID         int
	Created    time.Time
	Updated    time.Time
	OwnerID    int
	Course     AgencyCourse
	StartDate  time.Time
	EndDate    time.Time
	Operator   Operator
	Instructor Buddy
	Price      *Price
	Rating     *int
	Notes      string
}

func (c Certification) String() string {
	return fmt.Sprintf(
		"%s (%s to %s)",
		c.Course.Name,
		c.StartDate.Format(time.DateOnly),
		c.EndDate.Format(time.DateOnly),
	)
}

func (c Certification) Duration() time.Duration {
	return c.EndDate.Sub(c.StartDate)
}

type nullableCertification struct {
	ID         *int
	Created    *time.Time
	Updated    *time.Time
	OwnerID    *int
	Course     nullableAgencyCourse
	StartDate  *time.Time
	EndDate    *time.Time
	Operator   nullableOperator
	Instructor nullableBuddy
	Price      nullablePrice
	Rating     *int
	Notes      *string
}

func (nc nullableCertification) ToStruct() *Certification {
	if nc.ID == nil {
		return nil
	}

	return &Certification{
		ID:         *nc.ID,
		Created:    *nc.Created,
		Updated:    *nc.Updated,
		OwnerID:    *nc.OwnerID,
		Course:     *nc.Course.ToStruct(),
		StartDate:  *nc.StartDate,
		EndDate:    *nc.EndDate,
		Operator:   *nc.Operator.ToStruct(),
		Instructor: *nc.Instructor.ToStruct(),
		Price:      nc.Price.ToStruct(),
		Rating:     nc.Rating,
		Notes:      *nc.Notes,
	}
}

type CertificationModelInterface interface {
	Exists(id int) (bool, error)
	Insert(
		ownerID int,
		courseID int,
		startDate time.Time,
		endDate time.Time,
		operatorID int,
		instructorID int,
		priceAmount *float64,
		priceCurrencyID *int,
		rating *int,
		notes string,
	) (int, error)
	List(userID int, filters ListFilters) ([]Certification, PageData, error)
	ListAll(userID int) ([]Certification, error)
}

var certificationSelectQuery string = `
    select count(*) over(),
           ce.id, ce.created_at, ce.updated_at, ce.owner_id,
           ac.id,
           ag.id, ag.common_name, ag.full_name, ag.acronym, ag.url,
           ac.name, ac.url, ac.is_specialty_course, ac.is_tech_course,
           ac.is_pro_course,
           ce.start_date, ce.end_date,
           op.id, op.created_at, op.updated_at, op.owner_id,
           ot.id, ot.name, ot.description,
           op.name, op.street, op.suburb, op.state, op.postcode,
           oc.id, oc.name, oc.iso_number, oc.iso2_code,
           oc.iso3_code, oc.dialing_code, oc.capital,
           ou.id, ou.iso_alpha, ou.iso_number, ou.name, ou.exponent,
           op.website_url, op.email_address, op.phone_number, op.comments,
           bu.id, bu.version, bu.created_at, bu.updated_at, bu.owner_id,
           bu.name, bu.email, bu.phone_number,
           bu.agency_id, ba.common_name, ba.full_name, ba.acronym, ba.url,
           bu.agency_member_num, bu.notes,
           ce.price,
           cu.id, cu.iso_alpha, cu.iso_number, cu.name, cu.exponent,
           ce.rating, ce.notes
      from certifications ce
 left join agency_courses ac on ce.course_id = ac.id
 left join agencies       ag on ac.agency_id = ag.id
 left join operators      op on ce.operator_id = op.id
 left join operator_types ot on op.operator_type_id = ot.id
 left join countries      oc on op.country_id = oc.id
 left join currencies     ou on oc.currency_id = ou.id
 left join buddies        bu on ce.instructor_id = bu.id
 left join agencies       ba on bu.agency_id = ba.id
 left join currencies     cu on ce.currency_id = cu.id
     where ce.owner_id = $1
`

func certificationFromDBRow(rs RowScanner, totalRecords *int, ce *Certification) error {
	pr := &nullablePrice{}
	ia := &nullableAgency{}

	err := rs.Scan(
		totalRecords,
		&ce.ID,
		&ce.Created,
		&ce.Updated,
		&ce.OwnerID,
		&ce.Course.ID,
		&ce.Course.Agency.ID,
		&ce.Course.Agency.CommonName,
		&ce.Course.Agency.FullName,
		&ce.Course.Agency.Acronym,
		&ce.Course.Agency.URL,
		&ce.Course.Name,
		&ce.Course.URL,
		&ce.Course.IsSpecialtyCourse,
		&ce.Course.IsTechCourse,
		&ce.Course.IsProCourse,
		&ce.StartDate,
		&ce.EndDate,
		&ce.Operator.ID,
		&ce.Operator.Created,
		&ce.Operator.Updated,
		&ce.Operator.OwnerID,
		&ce.Operator.OperatorType.ID,
		&ce.Operator.OperatorType.Name,
		&ce.Operator.OperatorType.Description,
		&ce.Operator.Name,
		&ce.Operator.Street,
		&ce.Operator.Suburb,
		&ce.Operator.State,
		&ce.Operator.Postcode,
		&ce.Operator.Country.ID,
		&ce.Operator.Country.Name,
		&ce.Operator.Country.ISONumber,
		&ce.Operator.Country.ISO2Code,
		&ce.Operator.Country.ISO3Code,
		&ce.Operator.Country.DialingCode,
		&ce.Operator.Country.Capital,
		&ce.Operator.Country.Currency.ID,
		&ce.Operator.Country.Currency.ISOAlpha,
		&ce.Operator.Country.Currency.ISONumber,
		&ce.Operator.Country.Currency.Name,
		&ce.Operator.Country.Currency.Exponent,
		&ce.Operator.WebsiteURL,
		&ce.Operator.EmailAddress,
		&ce.Operator.PhoneNumber,
		&ce.Operator.Comments,
		&ce.Instructor.ID,
		&ce.Instructor.Version,
		&ce.Instructor.Created,
		&ce.Instructor.Updated,
		&ce.Instructor.OwnerID,
		&ce.Instructor.Name,
		&ce.Instructor.Email,
		&ce.Instructor.PhoneNumber,
		&ia.ID,
		&ia.CommonName,
		&ia.FullName,
		&ia.Acronym,
		&ia.URL,
		&ce.Instructor.AgencyMemberNum,
		&ce.Instructor.Notes,
		&pr.Amount,
		&pr.Currency.ID,
		&pr.Currency.ISOAlpha,
		&pr.Currency.ISONumber,
		&pr.Currency.Name,
		&pr.Currency.Exponent,
		&ce.Rating,
		&ce.Notes,
	)

	if err != nil {
		return err
	}

	ce.Price = pr.ToStruct()
	ce.Instructor.Agency = ia.ToStruct()

	return nil
}

type CertificationModel struct {
	DB *sql.DB
}

func (m *CertificationModel) Exists(id int) (bool, error) {
	return idExistsInTable(m.DB, id, "certifications", "id")
}

func (m *CertificationModel) Insert(
	ownerID int,
	courseID int,
	startDate time.Time,
	endDate time.Time,
	operatorID int,
	instructorID int,
	priceAmount *float64,
	priceCurrencyID *int,
	rating *int,
	notes string,
) (int, error) {
	stmt := `
        insert into certifications (
            owner_id, course_id, start_date, end_date, operator_id,
            instructor_id, price, currency_id, rating, notes
        ) values (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
        )
        returning id
    `

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result := m.DB.QueryRowContext(
		ctx,
		stmt,
		ownerID,
		courseID,
		startDate,
		endDate,
		operatorID,
		instructorID,
		priceAmount,
		priceCurrencyID,
		rating,
		notes,
	)

	var id int
	err := result.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *CertificationModel) List(
	userID int,
	filters ListFilters,
) ([]Certification, PageData, error) {
	limit := filters.limit()
	offset := filters.offset()
	stmt := fmt.Sprintf("%s limit $1 offset $2", certificationSelectQuery)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt, limit, offset)
	if err != nil {
		return nil, PageData{}, err
	}
	defer rows.Close()

	var totalRecords int
	records := []Certification{}
	for rows.Next() {
		var record Certification
		err := certificationFromDBRow(rows, &totalRecords, &record)
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

func (m *CertificationModel) ListAll(userID int) ([]Certification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, certificationSelectQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalRecords int
	var records []Certification
	for rows.Next() {
		var record Certification
		err := certificationFromDBRow(rows, &totalRecords, &record)
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
