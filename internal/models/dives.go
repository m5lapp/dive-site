package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Dive struct {
	ID                int
	Version           int
	Created           time.Time
	Updated           time.Time
	OwnerID           int
	Number            int
	Activity          string
	DiveSite          DiveSite
	Operator          *Operator
	Price             *Price
	Trip              *Trip
	Certification     *Certification
	DateTimeIn        time.Time
	MaxDepth          float64
	AvgDepth          *float64
	BottomTime        int
	SafetyStop        *int
	WaterTemp         *int
	AirTemp           *int
	Visibility        *float64
	Current           *Current
	Waves             *Waves
	Buddy             *Buddy
	BuddyRole         *BuddyRole
	Weight            *float64
	WeightNotes       string
	Equipment         []Equipment
	EquipmentNotes    string
	TankConfiguration TankConfiguration
	TankMaterial      TankMaterial
	TankVolume        float64
	GasMix            GasMix
	FO2               float64
	PressureIn        *int
	PressureOut       *int
	GasMixNotes       string
	EntryPoint        EntryPoint
	Properties        []DiveProperty
	Rating            *int
	Notes             string
}

type DiveModelInterface interface {
	Insert(
		ownerID int,
		number int,
		activity string,
		diveSiteID int,
		operatorID *int,
		priceAmount *float64,
		priceCurrencyID *int,
		tripID *int,
		certificationID *int,
		dateTimeIn time.Time,
		maxDepth float64,
		avgDepth *float64,
		bottomTime int,
		safetyStop *int,
		waterTemp *int,
		airTemp *int,
		visibility *float64,
		currentID *int,
		wavesID *int,
		buddyID *int,
		buddyRoleID *int,
		weight *float64,
		weightNotes string,
		equipmentIDs []int,
		equipmentNotes string,
		tankConfigurationID int,
		tankMaterialID int,
		tankVolume float64,
		gasMixID int,
		fo2 float64,
		pressureIn *int,
		pressureOut *int,
		gasMixNotes string,
		entryPointID int,
		propertyIDs []int,
		rating *int,
		notes string,
	) (int, error)
	List(userID int, filters ListFilters) ([]Dive, PageData, error)
}

var diveSelectQuery string = `
    select count(*) over(),
           -- dv.date_time_in - lag(
           --    dv.date_time_in, 1
           -- ) over (order by dv.date_time_in) surface_interval
           tr.id, tr.created_at, tr.updated_at, tr.owner_id, tr.name,
           tr.start_date, tr.end_date, tr.description, tr.rating
           op.id, op.created_at, op.updated_at, op.owner_id,
           ot.id, ot.name, ot.description,
           op.name, op.street, op.suburb, op.state, op.postcode,
           oc.id, oc.name, oc.iso_number, oc.iso2_code,
           oc.iso3_code, oc.dialing_code, oc.capital,
           ou.id, ou.iso_alpha, ou.iso_number, ou.name, ou.exponent,
           op.website_url, op.email_address, op.phone_number, op.comments
           tr.price,
           cu.id, cu.iso_alpha, cu.iso_number, cu.name, cu.exponent,
           tr.notes
      from trips tr
 left join operators      op on tr.operator_id = op.id
 left join operator_types ot on op.operator_type_id = op.id
 left join countries      oc on op.country_id = oc.id
 left join currencies     ou on oc.currency_id = ou.id
 left join currencies     cu on tr.currency_id = cu.id
     where tr.owner_id = $1
`

func diveFromDBRow(rs RowScanner, totalRecords *int, dv *Dive) error {
	op := nullableOperator{}
	pr := nullablePrice{}
	tr := nullableTrip{}
	ce := nullableCertification{}
	cu := nullableStaticDataItem{}
	wv := nullableStaticDataItem{}
	bu := nullableBuddy{}
	br := nullableBuddyRole{}
	tc := nullableStaticDataItem{}
	tm := nullableStaticDataItem{}
	gm := nullableStaticDataItem{}
	ep := nullableStaticDataItem{}

	err := rs.Scan(
		totalRecords,
		&dv.ID,
		&dv.Version,
		&dv.Created,
		&dv.Updated,
		&dv.OwnerID,
		&dv.Number,
		&dv.Activity,
		&dv.DiveSite.ID,
		&dv.DiveSite.Version,
		&dv.DiveSite.Created,
		&dv.DiveSite.Updated,
		&dv.DiveSite.OwnerId,
		&dv.DiveSite.Name,
		&dv.DiveSite.AltName,
		&dv.DiveSite.Location,
		&dv.DiveSite.Region,
		&dv.DiveSite.TimeZone,
		&dv.DiveSite.Latitude,
		&dv.DiveSite.Longitude,
		&dv.DiveSite.Altitude,
		&dv.DiveSite.MaxDepth,
		&dv.DiveSite.Notes,
		&dv.DiveSite.Rating,
		&dv.DiveSite.Country.ID,
		&dv.DiveSite.Country.Name,
		&dv.DiveSite.Country.ISONumber,
		&dv.DiveSite.Country.ISO2Code,
		&dv.DiveSite.Country.ISO3Code,
		&dv.DiveSite.Country.DialingCode,
		&dv.DiveSite.Country.Capital,
		&dv.DiveSite.Country.Currency.ID,
		&dv.DiveSite.Country.Currency.ISOAlpha,
		&dv.DiveSite.Country.Currency.ISONumber,
		&dv.DiveSite.Country.Currency.Name,
		&dv.DiveSite.Country.Currency.Exponent,
		&dv.DiveSite.WaterBody.ID,
		&dv.DiveSite.WaterBody.Name,
		&dv.DiveSite.WaterBody.Description,
		&dv.DiveSite.WaterType.ID,
		&dv.DiveSite.WaterType.Name,
		&dv.DiveSite.WaterType.Description,
		&dv.DiveSite.WaterType.Density,

		// Operator.
		&op.ID,
		&op.Created,
		&op.Updated,
		&op.OwnerID,
		&op.OperatorType.ID,
		&op.OperatorType.Name,
		&op.OperatorType.Description,
		&op.Name,
		&op.Street,
		&op.Suburb,
		&op.State,
		&op.Postcode,
		&op.Country.ID,
		&op.Country.Name,
		&op.Country.ISONumber,
		&op.Country.ISO2Code,
		&op.Country.ISO3Code,
		&op.Country.DialingCode,
		&op.Country.Capital,
		&op.Country.Currency.ID,
		&op.Country.Currency.ISOAlpha,
		&op.Country.Currency.ISONumber,
		&op.Country.Currency.Name,
		&op.Country.Currency.Exponent,

		// Dive price.
		&pr.Amount,
		&pr.Currency.ID,
		&pr.Currency.ISOAlpha,
		&pr.Currency.ISONumber,
		&pr.Currency.Name,
		&pr.Currency.Exponent,

		// Trip.
		&tr.ID,
		&tr.Created,
		&tr.Updated,
		&tr.OwnerID,
		&tr.Name,
		&tr.StartDate,
		&tr.EndDate,
		&tr.Description,
		&tr.Rating,
		&tr.Operator.ID,
		&tr.Operator.Created,
		&tr.Operator.Updated,
		&tr.Operator.OwnerID,
		&tr.Operator.OperatorType.ID,
		&tr.Operator.OperatorType.Name,
		&tr.Operator.OperatorType.Description,
		&tr.Operator.Name,
		&tr.Operator.Street,
		&tr.Operator.Suburb,
		&tr.Operator.State,
		&tr.Operator.Postcode,
		&tr.Operator.Country.ID,
		&tr.Operator.Country.Name,
		&tr.Operator.Country.ISONumber,
		&tr.Operator.Country.ISO2Code,
		&tr.Operator.Country.ISO3Code,
		&tr.Operator.Country.DialingCode,
		&tr.Operator.Country.Capital,
		&tr.Operator.Country.Currency.ID,
		&tr.Operator.Country.Currency.ISOAlpha,
		&tr.Operator.Country.Currency.ISONumber,
		&tr.Operator.Country.Currency.Name,
		&tr.Operator.Country.Currency.Exponent,
		&tr.Price.Amount,
		&tr.Price.Currency.ID,
		&tr.Price.Currency.ISOAlpha,
		&tr.Price.Currency.ISONumber,
		&tr.Price.Currency.Name,
		&tr.Price.Currency.Exponent,
		&tr.Notes,

		// Certification.
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
		&ce.Instructor.ID,
		&ce.Instructor.Version,
		&ce.Instructor.Created,
		&ce.Instructor.Updated,
		&ce.Instructor.OwnerID,
		&ce.Instructor.Name,
		&ce.Instructor.Email,
		&ce.Instructor.PhoneNumber,
		&ce.Instructor.Agency.ID,
		&ce.Instructor.Agency.CommonName,
		&ce.Instructor.Agency.FullName,
		&ce.Instructor.Agency.Acronym,
		&ce.Instructor.Agency.URL,
		&ce.Instructor.AgencyMemberNum,
		&ce.Instructor.Notes,
		&ce.Price.Amount,
		&ce.Price.Currency.ID,
		&ce.Price.Currency.ISOAlpha,
		&ce.Price.Currency.ISONumber,
		&ce.Price.Currency.Name,
		&ce.Price.Currency.Exponent,
		&ce.Rating,
		&ce.Notes,

		&dv.DateTimeIn,
		&dv.MaxDepth,
		&dv.AvgDepth,
		&dv.BottomTime,
		&dv.SafetyStop,
		&dv.WaterTemp,
		&dv.AirTemp,
		&dv.Visibility,

		// Current.
		&cu.ID,
		&cu.Sort,
		&cu.IsDefault,
		&cu.Name,
		&cu.Description,

		// Waves.
		&wv.ID,
		&wv.Sort,
		&wv.IsDefault,
		&wv.Name,
		&wv.Description,

		// Buddy.
		&bu.ID,
		&bu.Version,
		&bu.Created,
		&bu.Updated,
		&bu.OwnerID,
		&bu.Name,
		&bu.Email,
		&bu.PhoneNumber,
		&bu.Agency.ID,
		&bu.Agency.CommonName,
		&bu.Agency.FullName,
		&bu.Agency.Acronym,
		&bu.Agency.URL,
		&bu.AgencyMemberNum,
		&bu.Notes,

		// Buddu role.
		&br.ID,
		&br.Name,
		&br.Description,

		&dv.Weight,
		&dv.WeightNotes,
		&dv.EquipmentNotes,

		// Tank configuration.
		&tc.ID,
		&tc.Sort,
		&tc.IsDefault,
		&tc.Name,
		&tc.Description,

		// Tank material.
		&tm.ID,
		&tm.Sort,
		&tm.IsDefault,
		&tm.Name,
		&tm.Description,

		&dv.TankVolume,

		// Gas mix.
		&gm.ID,
		&gm.Sort,
		&gm.IsDefault,
		&gm.Name,
		&gm.Description,

		&dv.FO2,
		&dv.PressureIn,
		&dv.PressureOut,
		&dv.GasMixNotes,

		// Entry point.
		&ep.ID,
		&ep.Sort,
		&ep.IsDefault,
		&ep.Name,
		&ep.Description,

		&dv.Rating,
		&dv.Notes,
	)

	if err != nil {
		return err
	}

	dv.Operator = op.ToStruct()
	dv.Price = pr.ToStruct()
	dv.Current = nullableStaticDataItemToStruct[Current](cu)
	dv.Waves = nullableStaticDataItemToStruct[Waves](wv)
	dv.TankConfiguration = *nullableStaticDataItemToStruct[TankConfiguration](tc)
	dv.TankMaterial = *nullableStaticDataItemToStruct[TankMaterial](tm)
	dv.GasMix = *nullableStaticDataItemToStruct[GasMix](gm)
	dv.EntryPoint = *nullableStaticDataItemToStruct[EntryPoint](ep)

	return nil
}

type DiveModel struct {
	DB *sql.DB
}

// adjustDiveTimeZone takes a time.Time d which can be in any time.Location and
// adjusts it so that it represents the same time (i.e. without adjusting the
// clock value), but in the Location of the DiveSite ID given in siteID, then
// converted to UTC for storing in a database.
func (m *DiveModel) adjustDiveTimeZone(
	ctx context.Context,
	d time.Time,
	siteID int,
) (time.Time, error) {
	var siteTZStr string
	siteTZQuery := "select timezone from dive_sites where id = $1"

	err := m.DB.QueryRowContext(ctx, siteTZQuery, siteID).Scan(&siteTZStr)
	if err != nil {
		errMsg := "failed to get dive site (id %d) time zone: %w"
		return time.Time{}, fmt.Errorf(errMsg, siteID, err)
	}

	siteTZ, err := time.LoadLocation(siteTZStr)
	if err != nil {
		errMsg := "failed to load dive site (id %d) time zone (%s): %w"
		return time.Time{}, fmt.Errorf(errMsg, siteID, siteTZStr, err)
	}

	adjustedDate := time.Date(
		d.Year(),
		d.Month(),
		d.Day(),
		d.Hour(),
		d.Minute(),
		d.Second(),
		d.Nanosecond(),
		siteTZ,
	)
	return adjustedDate.UTC(), err
}

func (m *DiveModel) Insert(
	ownerID int,
	number int,
	activity string,
	diveSiteID int,
	operatorID *int,
	priceAmount *float64,
	priceCurrencyID *int,
	tripID *int,
	certificationID *int,
	dateTimeIn time.Time,
	maxDepth float64,
	avgDepth *float64,
	bottomTime int,
	safetyStop *int,
	waterTemp *int,
	airTemp *int,
	visibility *float64,
	currentID *int,
	wavesID *int,
	buddyID *int,
	buddyRoleID *int,
	weight *float64,
	weightNotes string,
	equipmentIDs []int,
	equipmentNotes string,
	tankConfigurationID int,
	tankMaterialID int,
	tankVolume float64,
	gasMixID int,
	fo2 float64,
	pressureIn *int,
	pressureOut *int,
	gasMixNotes string,
	entryPointID int,
	propertyIDs []int,
	rating *int,
	notes string,
) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Adjust the dateTimeIn so that it is in the same Location as the
	// diveSiteID and converted to UTC.
	dateTimeIn, err := m.adjustDiveTimeZone(ctx, dateTimeIn, diveSiteID)
	if err != nil {
		return 0, err
	}

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to start db transaction: %w", err)
	}

	stmt := `
        insert into dives (
            owner_id, number, activity, dive_site_id, operator_id, price,
            currency_id, trip_id, certification_id, date_time_in, max_depth,
            avg_depth, bottom_time, safety_stop, water_temp, air_temp,
            visibility, current_id, waves_id, buddy_id, buddy_role_id,
            weight_used, weight_notes, equipment_notes, tank_configuration_id,
            tank_material_id, tank_volume, gas_mix_id, fo2, pressure_in,
            pressure_out, gas_mix_notes, entry_point_id, rating, notes
        ) values (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
            $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28,
            $29, $30, $31, $32, $33, $34, $35
        )
        returning id
    `

	result := tx.QueryRowContext(
		ctx,
		stmt,
		ownerID,
		number,
		activity,
		diveSiteID,
		operatorID,
		priceAmount,
		priceCurrencyID,
		tripID,
		certificationID,
		dateTimeIn,
		maxDepth,
		avgDepth,
		bottomTime,
		safetyStop,
		waterTemp,
		airTemp,
		visibility,
		currentID,
		wavesID,
		buddyID,
		buddyRoleID,
		weight,
		weightNotes,
		equipmentNotes,
		tankConfigurationID,
		tankMaterialID,
		tankVolume,
		gasMixID,
		fo2,
		pressureIn,
		pressureOut,
		gasMixNotes,
		entryPointID,
		rating,
		notes,
	)

	var diveID int
	err = result.Scan(&diveID)
	if err != nil {
		_ = tx.Rollback()
		switch err.Error() {
		case `pq: duplicate key value violates unique constraint "dives_owner_id_number_key"`:
			return 0, ErrDuplicateDiveNumber
		default:
			return 0, err
		}
	}

	if len(equipmentIDs) > 0 {
		err = insertManyToManyIDs(
			ctx,
			tx,
			"dive_equipment",
			"dive_id",
			"equipment_id",
			diveID,
			equipmentIDs,
		)

		if err != nil {
			_ = tx.Rollback()
			return 0, err
		}
	}

	if len(propertyIDs) > 0 {
		err = insertManyToManyIDs(
			ctx,
			tx,
			"dive_dive_properties",
			"dive_id",
			"property_id",
			diveID,
			propertyIDs,
		)

		if err != nil {
			_ = tx.Rollback()
			return 0, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("failed to commit db transaction: %w", err)
	}

	return diveID, nil
}

func (m *DiveModel) List(userID int, filters ListFilters) ([]Dive, PageData, error) {
	limit := filters.limit()
	offset := filters.offset()
	stmt := fmt.Sprintf("%s limit $1 offset $2", diveSelectQuery)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt, limit, offset)
	if err != nil {
		return nil, PageData{}, err
	}
	defer rows.Close()

	var totalRecords int
	records := []Dive{}
	for rows.Next() {
		var record Dive
		err := diveFromDBRow(rows, &totalRecords, &record)
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
