package models

import (
	"context"
	"database/sql"
	"errors"
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
	SurfaceInterval   *time.Duration
	MaxDepth          float64
	AvgDepth          *float64
	BottomTime        time.Duration
	SafetyStop        *time.Duration
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

func (d Dive) DateTimeOut() time.Time {
	return d.DateTimeIn.Add(d.BottomTime)
}

func (d Dive) PressureDelta() int {
	if d.PressureIn == nil || d.PressureOut == nil {
		return 0
	}

	return *d.PressureIn - *d.PressureOut
}

func (d Dive) GasUsed() float64 {
	if d.PressureDelta() == 0 {
		return 0.0
	}

	tankCount := 1.0

	switch d.TankConfiguration.Name {
	case "Single Tank":
		tankCount = 1.0
	case "Sidemount", "Twinset":
		tankCount = 2.0
	default:
		return 0.0
	}

	return tankCount * d.TankVolume * float64(d.PressureDelta())
}

// Calculates the diver's SAC (Surface Air Consumption) rate in litres per
// minute if AvgDepth and BottomTime are set and GasUsed can be calculated.
// Returns 0.0 if the value cannot be calculated.
func (d Dive) SACRate() float64 {
	gasUsed := d.GasUsed()

	if gasUsed == 0.0 || d.AvgDepth == nil {
		return 0.0
	}

	litresPerMinute := gasUsed / d.BottomTime.Minutes()
	avgPressure := (*d.AvgDepth + 1.0) / 10.0
	sacRate := litresPerMinute / avgPressure

	return sacRate
}

func (d Dive) IsAltitudeDive() bool {
	altitudeDive := d.DiveSite.Altitude >= 300

	requiresCorrection := d.DiveSite.Altitude >= 91 &&
		d.DiveSite.Altitude < 300 &&
		d.MaxDepth >= 44.0

	return altitudeDive || requiresCorrection
}

func (d Dive) IsDeepDive() bool {
	return d.MaxDepth > 30.0
}

func (d Dive) IsTrainingDive() bool {
	return d.Certification != nil
}

type DiveModelInterface interface {
	GetOneByID(ownerID, id int) (Dive, error)
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
		bottomTime time.Duration,
		safetyStop *time.Duration,
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
      with buddy_dive_stats as (
        select dv.buddy_id buddy_id, count(dv.id) dives_with,
               min(dv.date_time_in) first_dive_with,
               max(dv.date_time_in) last_dive_with
          from dives dv
         where dv.owner_id = $1
      group by dv.buddy_id
           )
    select count(*) over(),
           dv.id, dv.version, dv.created_at, dv.updated_at, dv.owner_id,
           dv.number, dv.activity,
           ds.id, ds.version, ds.created_at, ds.updated_at,
           ds.owner_id, ds.name, ds.alt_name, ds.location, ds.region,
           ds.timezone, ds.latitude, ds.longitude, ds.altitude, ds.max_depth,
           ds.notes, ds.rating, dsco.id, dsco.name, dsco.iso_number,
           dsco.iso2_code, dsco.iso3_code, dsco.dialing_code, dsco.capital,
           dscu.id, dscu.iso_alpha, dscu.iso_number, dscu.name, dscu.exponent,
           wb.id, wb.name, wb.description, wt.id, wt.name, wt.description,
           wt.density,
           op.id, op.created_at, op.updated_at, op.owner_id,
           opot.id, opot.name, opot.description,
           op.name, op.street, op.suburb, op.state, op.postcode,
           opco.id, opco.name, opco.iso_number, opco.iso2_code,
           opco.iso3_code, opco.dialing_code, opco.capital,
           opcu.id, opcu.iso_alpha, opcu.iso_number, opcu.name, opcu.exponent,
           op.website_url, op.email_address, op.phone_number, op.comments,
           dv.price, prcu.id, prcu.iso_alpha, prcu.iso_number, prcu.name,
           prcu.exponent,
           tr.id, tr.created_at, tr.updated_at, tr.owner_id, tr.name,
           tr.start_date, tr.end_date, tr.description, tr.rating,
           trop.id, trop.created_at, trop.updated_at, trop.owner_id,
           trot.id, trot.name, trot.description,
           trop.name, trop.street, trop.suburb, trop.state, trop.postcode,
           troc.id, troc.name, troc.iso_number, troc.iso2_code,
           troc.iso3_code, troc.dialing_code, troc.capital,
           trou.id, trou.iso_alpha, trou.iso_number, trou.name, trou.exponent,
           trop.website_url, trop.email_address, trop.phone_number,
           trop.comments,
           tr.price,
           trcu.id, trcu.iso_alpha, trcu.iso_number, trcu.name, trcu.exponent,
           tr.notes,
           ce.id, ce.created_at, ce.updated_at, ce.owner_id,
           ceac.id,
           ceag.id, ceag.common_name, ceag.full_name, ceag.acronym, ceag.url,
           ceac.name, ceac.url, ceac.is_specialty_course, ceac.is_tech_course,
           ceac.is_pro_course,
           ce.start_date, ce.end_date,
           ceop.id, ceop.created_at, ceop.updated_at, ceop.owner_id,
           ceot.id, ceot.name, ceot.description,
           ceop.name, ceop.street, ceop.suburb, ceop.state, ceop.postcode,
           ceoc.id, ceoc.name, ceoc.iso_number, ceoc.iso2_code,
           ceoc.iso3_code, ceoc.dialing_code, ceoc.capital,
           ceou.id, ceou.iso_alpha, ceou.iso_number, ceou.name, ceou.exponent,
           ceop.website_url, ceop.email_address, ceop.phone_number,
           ceop.comments,
           cebu.id, cebu.version, cebu.created_at, cebu.updated_at,
           cebu.owner_id, cebu.name, cebu.email, cebu.phone_number,
           ceba.id, ceba.common_name, ceba.full_name, ceba.acronym,
           ceba.url,
           cebu.agency_member_num,
           coalesce(cebs.dives_with, 0),
           cebs.first_dive_with, cebs.last_dive_with,
           cebu.notes,
           ce.price,
           cecu.id, cecu.iso_alpha, cecu.iso_number, cecu.name, cecu.exponent,
           ce.rating, ce.notes,
           dv.date_time_in,
           si.surface_interval,
           dv.max_depth, dv.avg_depth, dv.bottom_time, dv.safety_stop,
           dv.water_temp, dv.air_temp, dv.visibility,
           cu.id, cu.sort, cu.is_default, cu.name, cu.description,
           wv.id, wv.sort, wv.is_default, wv.name, wv.description,
           bu.id, bu.version, bu.created_at, bu.updated_at, bu.owner_id,
           bu.name, bu.email, bu.phone_number,
           buag.id, buag.common_name, buag.full_name, buag.acronym, buag.url,
           bu.agency_member_num,
           coalesce(cbds.dives_with, 0),
           cbds.first_dive_with, cbds.last_dive_with,
           bu.notes,
           br.id, br.name, br.description,
           dv.weight_used, dv.weight_notes, dv.equipment_notes,
           tc.id, tc.sort, tc.is_default, tc.name, tc.description,
           tm.id, tm.sort, tm.is_default, tm.name, tm.description,
           dv.tank_volume,
           gm.id, gm.sort, gm.is_default, gm.name, gm.description,
           dv.fo2, dv.pressure_in, dv.pressure_out, dv.gas_mix_notes,
           ep.id, ep.sort, ep.is_default, ep.name, ep.description,
           dv.rating, dv.notes
      from dives dv
inner join (
    -- Calculate the surface interval which we take to be the length of time
    -- since the last logged dive for that user. As the lag function is only run
    -- against the results within the query window, then in order to do this
    -- accurately, we need to run a full table subquery and get the result from
    -- there. The result is multiplied by 10^9 in order to get the value in
    -- nanoseconds.
    select id, (
        extract(epoch from age(
            date_time_in,
            lag(date_time_in + make_interval(secs => bottom_time / 10^9), 1) over (
                partition by owner_id order by date_time_in
        ))) * 10^9)::bigint surface_interval
      from dives             ) si   on dv.id = si.id
 left join dive_sites          ds   on dv.dive_site_id = ds.id
 left join countries           dsco on ds.country_id = dsco.id
 left join currencies          dscu on dsco.currency_id = dscu.id
 left join water_bodies        wb   on ds.water_body_id = wb.id
 left join water_types         wt   on ds.water_type_id = wt.id
 left join operators           op   on dv.operator_id = op.id
 left join operator_types      opot on op.operator_type_id = opot.id
 left join countries           opco on op.country_id = opco.id
 left join currencies          opcu on opco.currency_id = opcu.id
 left join currencies          prcu on dv.currency_id = prcu.id
 left join trips               tr   on dv.trip_id = tr.id
 left join operators           trop on tr.operator_id = trop.id
 left join operator_types      trot on trop.operator_type_id = trot.id
 left join countries           troc on trop.country_id = troc.id
 left join currencies          trou on troc.currency_id = trou.id
 left join currencies          trcu on tr.currency_id = trcu.id
 left join certifications      ce   on dv.certification_id = ce.id
 left join agency_courses      ceac on ce.course_id = ceac.id
 left join agencies            ceag on ceac.agency_id = ceag.id
 left join operators           ceop on ce.operator_id = ceop.id
 left join operator_types      ceot on ceop.operator_type_id = ceot.id
 left join countries           ceoc on ceop.country_id = ceoc.id
 left join currencies          ceou on ceoc.currency_id = ceou.id
 left join buddies             cebu on ce.instructor_id = cebu.id
 left join buddy_dive_stats    cbds on ce.instructor_id = cbds.buddy_id
 left join agencies            ceba on cebu.agency_id = ceba.id
 left join currencies          cecu on ce.currency_id = cecu.id
 left join currents            cu   on dv.current_id = cu.id
 left join waves               wv   on dv.waves_id = wv.id
 left join buddies             bu   on dv.buddy_id = bu.id
 left join buddy_dive_stats    buds on dv.buddy_id = buds.buddy_id
 left join agencies            buag on bu.agency_id = buag.id
 left join buddy_roles         br   on dv.buddy_role_id = br.id
 left join tank_configurations tc   on dv.tank_configuration_id = tc.id
 left join tank_materials      tm   on dv.tank_material_id = tm.id
 left join gas_mixes           gm   on dv.gas_mix_id = gm.id
 left join entry_points        ep   on dv.entry_point_id = ep.id
     where dv.owner_id = $1
`

func diveFromDBRow(rs RowScanner, totalRecords *int, dv *Dive) error {
	var bottomTimeNanos int64
	var safetyStopNanos *int64
	var surfaceIntervalNanos *int64
	op := nullableOperator{}
	pr := nullablePrice{}
	tr := nullableTrip{}
	ce := nullableCertification{}
	cu := nullableStaticDataItem{}
	wv := nullableStaticDataItem{}
	bu := nullableBuddy{}
	br := nullableBuddyRole{}

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
		&op.WebsiteURL,
		&op.EmailAddress,
		&op.PhoneNumber,
		&op.Comments,

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
		&tr.Operator.WebsiteURL,
		&tr.Operator.EmailAddress,
		&tr.Operator.PhoneNumber,
		&tr.Operator.Comments,
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
		&surfaceIntervalNanos,
		&dv.MaxDepth,
		&dv.AvgDepth,
		&bottomTimeNanos,
		&safetyStopNanos,
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
		&dv.TankConfiguration.ID,
		&dv.TankConfiguration.Sort,
		&dv.TankConfiguration.IsDefault,
		&dv.TankConfiguration.Name,
		&dv.TankConfiguration.Description,

		// Tank material.
		&dv.TankMaterial.ID,
		&dv.TankMaterial.Sort,
		&dv.TankMaterial.IsDefault,
		&dv.TankMaterial.Name,
		&dv.TankMaterial.Description,

		&dv.TankVolume,

		// Gas mix.
		&dv.GasMix.ID,
		&dv.GasMix.Sort,
		&dv.GasMix.IsDefault,
		&dv.GasMix.Name,
		&dv.GasMix.Description,

		&dv.FO2,
		&dv.PressureIn,
		&dv.PressureOut,
		&dv.GasMixNotes,

		// Entry point.
		&dv.EntryPoint.ID,
		&dv.EntryPoint.Sort,
		&dv.EntryPoint.IsDefault,
		&dv.EntryPoint.Name,
		&dv.EntryPoint.Description,

		&dv.Rating,
		&dv.Notes,
	)

	if err != nil {
		return err
	}

	dv.Operator = op.ToStruct()
	dv.Price = pr.ToStruct()
	dv.Trip = tr.ToStruct()
	dv.Certification = ce.ToStruct()
	dv.Current = nullableStaticDataItemToStruct[Current](cu)
	dv.Waves = nullableStaticDataItemToStruct[Waves](wv)
	dv.Buddy = bu.ToStruct()
	dv.BuddyRole = br.ToStruct()

	dv.BottomTime = time.Duration(bottomTimeNanos)

	// Adjust the Dive's DateTimeIn from UTC to the time zone of the dive site.
	dv.DateTimeIn = dv.DateTimeIn.In(&dv.DiveSite.TimeZone.Location)

	if safetyStopNanos != nil {
		ss := time.Duration(*safetyStopNanos)
		dv.SafetyStop = &ss
	}

	if surfaceIntervalNanos != nil {
		si := time.Duration(*surfaceIntervalNanos)
		dv.SurfaceInterval = &si
	}

	return nil
}

type DiveModel struct {
	DB             *sql.DB
	equipmentModel EquipmentModelInterface
	propertyModel  DivePropertyModelInterface
}

func NewDiveModel(
	db *sql.DB,
	equipmentModel EquipmentModelInterface,
	propertyModel DivePropertyModelInterface,
) (*DiveModel, error) {
	if db == nil {
		return nil, fmt.Errorf("diveModel db cannot be nil")
	}

	if equipmentModel == nil {
		return nil, fmt.Errorf("diveModel equipmentModel cannot be nil")
	}

	if propertyModel == nil {
		return nil, fmt.Errorf("diveModel propertyModel cannot be nil")
	}

	return &DiveModel{
		DB:             db,
		equipmentModel: equipmentModel,
		propertyModel:  propertyModel,
	}, nil
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

func (m *DiveModel) GetOneByID(ownerID, id int) (Dive, error) {
	stmt := fmt.Sprintf("%s and dv.id = $2", diveSelectQuery)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var totalRecords int
	var dive Dive
	row := m.DB.QueryRowContext(ctx, stmt, ownerID, id)
	err := diveFromDBRow(row, &totalRecords, &dive)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Dive{}, ErrNoRecord
		} else {
			return Dive{}, err
		}
	}

	equipment, err := m.equipmentModel.GetAllForDive(id)
	if err != nil {
		return Dive{}, err
	}

	properties, err := m.propertyModel.GetAllForDive(id)
	if err != nil {
		return Dive{}, err
	}

	dive.Equipment = equipment
	dive.Properties = properties

	return dive, nil
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
	bottomTime time.Duration,
	safetyStop *time.Duration,
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

	var safetyStopNanos *int64
	if safetyStop != nil {
		ss := safetyStop.Nanoseconds()
		safetyStopNanos = &ss
	}

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
		bottomTime.Nanoseconds(),
		safetyStopNanos,
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
	stmt := fmt.Sprintf("%s order by date_time_in desc limit $2 offset $3", diveSelectQuery)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt, userID, limit, offset)
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
