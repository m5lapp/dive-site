package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type aggregateDiveStats struct {
	Dives         int
	FirstDiveDate time.Time
	LastDiveDate  time.Time

	BottomTime struct {
		Avg time.Duration
		Max time.Duration
		Sum time.Duration
	}

	AvgDiveDepth struct {
		Avg float64
		Max float64
	}

	MaxDiveDepth struct {
		Avg float64
		Max float64
	}
}

type monthlyDiveStats struct {
	Month time.Time
	aggregateDiveStats
}

type countryDiveStats struct {
	Country   Country
	Continent string // TODO: This should be in the Country struct.
	aggregateDiveStats
}

type diveSiteDiveStats struct {
	DiveSite DiveSite
	aggregateDiveStats
}

type buddyDiveStats struct {
	Buddy Buddy
	aggregateDiveStats
}

type DiveStats struct {
	aggregateDiveStats
	DivesByMonth    []monthlyDiveStats
	DivesByCountry  []countryDiveStats
	DivesByDiveSite []diveSiteDiveStats
	DivesByBuddy    []buddyDiveStats
}

func (m DiveModel) GetDiveStats(userID int) (DiveStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := DiveStats{}

	err := m.getGeneralStats(ctx, userID, &stats)
	if err != nil {
		return DiveStats{}, fmt.Errorf("error getting general user dive stats: %s", err)
	}

	err = m.getStatsByMonth(ctx, userID, &stats)
	if err != nil {
		return DiveStats{}, fmt.Errorf("error getting monthly user dive stats: %s", err)
	}

	err = m.getStatsByCountry(ctx, userID, &stats)
	if err != nil {
		return DiveStats{}, fmt.Errorf("error getting country user dive stats: %s", err)
	}

	err = m.getStatsByDiveSite(ctx, userID, &stats)
	if err != nil {
		return DiveStats{}, fmt.Errorf("error getting dive site user dive stats: %s", err)
	}

	err = m.getStatsByBuddy(ctx, userID, &stats)
	if err != nil {
		return DiveStats{}, fmt.Errorf("error getting buddy user dive stats: %s", err)
	}

	return stats, nil
}

var aggregateFields string = `
    count(dv.id) dives,
    min(dv.date_time_in) first_dive_date, max(dv.date_time_in) last_dive_date,
    avg(dv.bottom_time) avg_bottom_time, max(dv.bottom_time) max_bottom_time,
                                         sum(dv.bottom_time) bottom_time_sum,
    coalesce(avg(dv.avg_depth), 0.0) avg_avg_depth,
    coalesce(max(dv.avg_depth), 0.0) max_avg_depth,
    avg(dv.max_depth) avg_max_depth, max(dv.max_depth) max_max_depth
`

func statsFromDBRow(rs RowScanner, stats *aggregateDiveStats, otherData ...any) error {
	var bottomTimeAvg, bottomTimeMax, bottomTimeSum float64

	destinations := []any{
		&stats.Dives,
		&stats.FirstDiveDate,
		&stats.LastDiveDate,
		&bottomTimeAvg,
		&bottomTimeMax,
		&bottomTimeSum,
		&stats.AvgDiveDepth.Avg,
		&stats.AvgDiveDepth.Max,
		&stats.MaxDiveDepth.Avg,
		&stats.MaxDiveDepth.Max,
	}

	// Merge the items from otherData into the slice of destinations.
	for _, item := range otherData {
		destinations = append(destinations, item)
	}

	err := rs.Scan(destinations...)
	if err != nil {
		return err
	}

	stats.BottomTime.Avg = time.Duration(bottomTimeAvg)
	stats.BottomTime.Max = time.Duration(bottomTimeMax)
	stats.BottomTime.Sum = time.Duration(bottomTimeSum)

	return nil
}

func (m DiveModel) getGeneralStats(
	ctx context.Context,
	userID int,
	stats *DiveStats,
) error {
	if stats == nil {
		return errors.New("nil UserDiveStats struct passed to getGeneralStats")
	}

	query := "select %s from dives dv where dv.owner_id = $1"
	stmt := fmt.Sprintf(query, aggregateFields)

	row := m.DB.QueryRowContext(ctx, stmt, userID)
	err := statsFromDBRow(row, &stats.aggregateDiveStats)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoRecord
		} else {
			return err
		}
	}

	return nil
}

func (m DiveModel) getStatsByMonth(
	ctx context.Context,
	userID int,
	stats *DiveStats,
) error {
	if stats == nil {
		return errors.New("nil UserDiveStats struct passed to getStatsByMonth")
	}

	query := `
        select %s, date_trunc('month', dv.date_time_in) as month
          from dives dv
         where dv.owner_id = $1
      group by month
      order by month desc
    `

	stmt := fmt.Sprintf(query, aggregateFields)

	rows, err := m.DB.QueryContext(ctx, stmt, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var record monthlyDiveStats
		err := statsFromDBRow(rows, &record.aggregateDiveStats, &record.Month)
		if err != nil {
			return err
		}
		stats.DivesByMonth = append(stats.DivesByMonth, record)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}

func (m DiveModel) getStatsByCountry(
	ctx context.Context,
	userID int,
	stats *DiveStats,
) error {
	if stats == nil {
		return errors.New("nil UserDiveStats struct passed to getStatsByMonth")
	}

	query := `
        select %[1]s, %[2]s
          from dives dv
     left join dive_sites ds on dv.dive_site_id = ds.id
     left join countries  co on ds.country_id = co.id
     left join currencies cu on co.currency_id = cu.id
         where dv.owner_id = $1
      group by %[2]s
      order by dives desc
         limit 10
    `

	countryColumns := `
        co.id, co.name, co.iso_number, co.iso2_code, co.iso3_code,
        co.dialing_code, co.capital,
        cu.id, cu.iso_alpha, cu.iso_number, cu.name, cu.exponent,
        co.continent
    `

	stmt := fmt.Sprintf(query, aggregateFields, countryColumns)

	rows, err := m.DB.QueryContext(ctx, stmt, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var record countryDiveStats
		countryFields := []any{
			&record.Country.ID,
			&record.Country.Name,
			&record.Country.ISONumber,
			&record.Country.ISO2Code,
			&record.Country.ISO3Code,
			&record.Country.DialingCode,
			&record.Country.Capital,
			&record.Country.Currency.ID,
			&record.Country.Currency.ISOAlpha,
			&record.Country.Currency.ISONumber,
			&record.Country.Currency.Name,
			&record.Country.Currency.Exponent,
			&record.Continent,
		}

		err := statsFromDBRow(rows, &record.aggregateDiveStats, countryFields...)
		if err != nil {
			return err
		}
		stats.DivesByCountry = append(stats.DivesByCountry, record)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}

func (m DiveModel) getStatsByDiveSite(
	ctx context.Context,
	userID int,
	stats *DiveStats,
) error {
	if stats == nil {
		return errors.New("nil UserDiveStats struct passed to getStatsByDiveSite")
	}

	query := `
      with dive_site_dive_stats as (
        select dv.dive_site_id dive_site_id,
               count(dv.id) dives_at,
               min(dv.date_time_in) first_dive_at,
               max(dv.date_time_in) last_dive_at
          from dives dv
         where dv.owner_id = $1
      group by dv.dive_site_id
           )
    select %[1]s, %[2]s
      from dives dv
 left join dive_sites           ds on dv.dive_site_id = ds.id
 left join dive_site_dive_stats st on ds.id = st.dive_site_id
 left join countries            co on ds.country_id = co.id
 left join currencies           cu on co.currency_id = cu.id
 left join water_bodies         wb on ds.water_body_id = wb.id
 left join water_types          wt on ds.water_type_id = wt.id
     where dv.owner_id = $1
  group by %[2]s
  order by dives desc
     limit 10
    `

	diveSiteColumns := `
        ds.id, ds.version, ds.created_at, ds.updated_at,
        ds.owner_id,
        coalesce(st.dives_at, 0), st.first_dive_at, st.last_dive_at,
        ds.name, ds.alt_name, ds.location, ds.region,
        ds.timezone, ds.latitude, ds.longitude, ds.altitude, ds.max_depth,
        ds.notes, ds.rating, co.id, co.name, co.iso_number, co.iso2_code,
        co.iso3_code, co.dialing_code, co.capital, cu.id, cu.iso_alpha,
        cu.iso_number, cu.name, cu.exponent, wb.id, wb.name, wb.description,
        wt.id, wt.name, wt.description, wt.density
    `

	stmt := fmt.Sprintf(query, aggregateFields, diveSiteColumns)

	rows, err := m.DB.QueryContext(ctx, stmt, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var record diveSiteDiveStats
		diveSiteFields := []any{
			&record.DiveSite.ID,
			&record.DiveSite.Version,
			&record.DiveSite.Created,
			&record.DiveSite.Updated,
			&record.DiveSite.OwnerId,
			&record.DiveSite.DivesAt,
			&record.DiveSite.FirstDiveAt,
			&record.DiveSite.LastDiveAt,
			&record.DiveSite.Name,
			&record.DiveSite.AltName,
			&record.DiveSite.Location,
			&record.DiveSite.Region,
			&record.DiveSite.TimeZone,
			&record.DiveSite.Latitude,
			&record.DiveSite.Longitude,
			&record.DiveSite.Altitude,
			&record.DiveSite.MaxDepth,
			&record.DiveSite.Notes,
			&record.DiveSite.Rating,
			&record.DiveSite.Country.ID,
			&record.DiveSite.Country.Name,
			&record.DiveSite.Country.ISONumber,
			&record.DiveSite.Country.ISO2Code,
			&record.DiveSite.Country.ISO3Code,
			&record.DiveSite.Country.DialingCode,
			&record.DiveSite.Country.Capital,
			&record.DiveSite.Country.Currency.ID,
			&record.DiveSite.Country.Currency.ISOAlpha,
			&record.DiveSite.Country.Currency.ISONumber,
			&record.DiveSite.Country.Currency.Name,
			&record.DiveSite.Country.Currency.Exponent,
			&record.DiveSite.WaterBody.ID,
			&record.DiveSite.WaterBody.Name,
			&record.DiveSite.WaterBody.Description,
			&record.DiveSite.WaterType.ID,
			&record.DiveSite.WaterType.Name,
			&record.DiveSite.WaterType.Description,
			&record.DiveSite.WaterType.Density,
		}

		err := statsFromDBRow(rows, &record.aggregateDiveStats, diveSiteFields...)
		if err != nil {
			return err
		}
		stats.DivesByDiveSite = append(stats.DivesByDiveSite, record)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}

func (m DiveModel) getStatsByBuddy(
	ctx context.Context,
	userID int,
	stats *DiveStats,
) error {
	if stats == nil {
		return errors.New("nil UserDiveStats struct passed to getStatsByBuddy")
	}

	query := `
      with buddy_dive_stats as (
        select dv.buddy_id buddy_id, count(dv.id) dives_with,
               min(dv.date_time_in) first_dive_with,
               max(dv.date_time_in) last_dive_with
          from dives dv
         where dv.owner_id = $1
      group by dv.buddy_id
           )
    select %[1]s, %[2]s
      from dives dv
-- Inner join here as buddies can be null.
inner join buddies          bu on dv.buddy_id = bu.id
 left join agencies         ag on bu.agency_id = ag.id
 left join buddy_dive_stats ds on bu.id = ds.buddy_id
     where dv.owner_id = $1
  group by %[2]s
  order by dives desc
     limit 10
    `

	buddyColumns := `
        bu.id, bu.version, bu.created_at, bu.updated_at, bu.owner_id,
        bu.name, bu.email, bu.phone_number,
        bu.agency_id, ag.common_name, ag.full_name, ag.acronym, ag.url,
        bu.agency_member_num,
        coalesce(ds.dives_with, 0),
        ds.first_dive_with, ds.last_dive_with,
        bu.notes
    `

	stmt := fmt.Sprintf(query, aggregateFields, buddyColumns)

	rows, err := m.DB.QueryContext(ctx, stmt, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	ag := &nullableAgency{}

	for rows.Next() {
		var record buddyDiveStats
		diveSiteFields := []any{
			&record.Buddy.ID,
			&record.Buddy.Version,
			&record.Buddy.Created,
			&record.Buddy.Updated,
			&record.Buddy.OwnerID,
			&record.Buddy.Name,
			&record.Buddy.Email,
			&record.Buddy.PhoneNumber,
			&ag.ID,
			&ag.CommonName,
			&ag.FullName,
			&ag.Acronym,
			&ag.URL,
			&record.Buddy.AgencyMemberNum,
			&record.Buddy.DivesWith,
			&record.Buddy.FirstDiveWith,
			&record.Buddy.LastDiveWith,
			&record.Buddy.Notes,
		}

		err := statsFromDBRow(rows, &record.aggregateDiveStats, diveSiteFields...)
		if err != nil {
			return err
		}
		record.Buddy.Agency = ag.ToStruct()
		stats.DivesByBuddy = append(stats.DivesByBuddy, record)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}
