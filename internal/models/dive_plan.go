package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/m5lapp/diveplanner"
	"github.com/m5lapp/diveplanner/gasmix"
)

type DivePlan struct {
	ID      int
	Version int
	OwnerId int
	diveplanner.DivePlan
}

func (dp *DivePlan) String() string {
	return fmt.Sprintf(
		"%s - %.0fmin@%.0fm, %s",
		dp.Name,
		dp.Runtime(),
		dp.MaxDepth(),
		dp.GasMix.MixType(),
	)
}

func (dp *DivePlan) ChartProfileData(resolution int) (map[string]template.JS, error) {
	profile := dp.ChartProfile(resolution)

	var data map[string]template.JS = make(map[string]template.JS)
	var times []int
	var depths []float64
	var ndls []int
	var gas []float64

	for _, sample := range profile {
		times = append(times, sample.Time)
		depths = append(depths, sample.Depth)
		ndls = append(ndls, sample.NDL)
		gas = append(gas, sample.Gas)
	}

	timesJSON, err := json.Marshal(times)
	if err != nil {
		return data, err
	}

	depthsJSON, err := json.Marshal(depths)
	if err != nil {
		return data, err
	}

	gasJSON, err := json.Marshal(gas)
	if err != nil {
		return data, err
	}

	ndlsJSON, err := json.Marshal(ndls)
	if err != nil {
		return data, err
	}

	data["times"] = template.JS(timesJSON)
	data["depths"] = template.JS(depthsJSON)
	data["ndls"] = template.JS(ndlsJSON)
	data["gas"] = template.JS(gasJSON)

	return data, nil
}

type DivePlanModelInterface interface {
	Insert(
		ownerID int,
		name string,
		notes string,
		isSoloDive bool,
		descentRate float64,
		ascentRate float64,
		sacRate float64,
		tankCount int,
		tankVolume float64,
		workingPressure int,
		diveFactor float64,
		fn2 float64,
		fhe float64,
		maxPPO2 float64,
		stops []DivePlanStopInput,
	) (int, error)

	Update(
		id int,
		ownerID int,
		name string,
		notes string,
		isSoloDive bool,
		descentRate float64,
		ascentRate float64,
		sacRate float64,
		tankCount int,
		tankVolume float64,
		workingPressure int,
		diveFactor float64,
		fn2 float64,
		fhe float64,
		maxPPO2 float64,
		stops []DivePlanStopInput,
	) error

	GetOneByID(id, diverID int) (DivePlan, error)

	List(diverID int, ListControls Pager, sort []SortDivePlan) ([]DivePlan, PageData, error)

	Exists(id int) (bool, error)
}

var divePlanSelectQuery string = `
    select count(*) over(), dp.id, dp.version, dp.created_at, dp.updated_at,
           dp.owner_id, dp.name, dp.notes, dp.is_solo_dive, dp.descent_rate,
           dp.ascent_rate, dp.sac_rate, dp.tank_count, dp.tank_volume,
           dp.working_pressure, dp.dive_factor, dp.fn2, dp.fhe, dp.max_ppo2,
           coalesce(
               jsonb_agg(
                   jsonb_build_object(
                       'depth',         ds.depth,
                       'duration',      ds.duration,
                       'is_transition', ds.is_transition,
                       'comment',       ds.comment
                   ) order by ds.sort
               ) filter (where ds.id is not null),
               '[]'::jsonb
           ) as stops_json
      from dive_plans dp
      left join dive_plan_stops ds on dp.id = ds.dive_plan_id
     where dp.owner_id = $1
       and (dp.id = $2 or $2::bigint is null)
     group by dp.id, dp.version, dp.created_at, dp.updated_at,
           dp.owner_id, dp.name, dp.notes, dp.is_solo_dive, dp.descent_rate,
           dp.ascent_rate, dp.sac_rate, dp.tank_count, dp.tank_volume,
           dp.working_pressure, dp.dive_factor, dp.fn2, dp.fhe, dp.max_ppo2
`

func divePlanFromDBRow(rs RowScanner, totalRecords *int, dp *DivePlan) error {
	var fN2, fHe float64
	var stopsRaw []byte

	err := rs.Scan(
		totalRecords,
		&dp.ID,
		&dp.Version,
		&dp.Created,
		&dp.Updated,
		&dp.OwnerId,
		&dp.Name,
		&dp.Notes,
		&dp.IsSoloDive,
		&dp.DescentRate,
		&dp.AscentRate,
		&dp.SACRate,
		&dp.TankCount,
		&dp.TankCapacity,
		&dp.WorkingPressure,
		&dp.DiveFactor,
		&fN2,
		&fHe,
		&dp.MaxPPO2,
		&stopsRaw,
	)

	if err != nil {
		return err
	}

	dp.GasMix = &gasmix.GasMix{
		FO2: 1.0 - (fN2 + fHe),
		FN2: fN2,
		FHe: fHe,
	}

	err = json.Unmarshal(stopsRaw, &dp.Stops)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stops for dive plan %d: %w", dp.ID, err)
	}

	return nil
}

type DivePlanModel struct {
	DB *sql.DB
}

type DivePlanStopInput struct {
	Depth    float64
	Duration float64
	Comment  string
}

func insertDivePlanStops(
	ctx context.Context,
	db sqlExecer,
	deleteBeforeInsert bool,
	divePlanID int,
	stops []DivePlanStopInput,
) error {
	if deleteBeforeInsert {
		stopDeleteStmt := `delete from dive_plan_stops where dive_plan_id = $1`
		_, err := db.ExecContext(ctx, stopDeleteStmt, divePlanID)
		if err != nil {
			return err
		}
	}

	stopStmt := strings.Builder{}
	stopStmt.WriteString("insert into dive_plan_stops (")
	stopStmt.WriteString("sort, dive_plan_id, depth, duration, comment")
	stopStmt.WriteString(") values ")

	var params []any
	for i, stop := range stops {
		j := i * 5

		if i > 0 {
			stopStmt.WriteString(",")
		}

		stopStmt.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", j+1, j+2, j+3, j+4, j+5))
		params = append(params, i, divePlanID, stop.Depth, stop.Duration, stop.Comment)
	}

	_, err := db.ExecContext(
		ctx,
		stopStmt.String(),
		params...,
	)

	return err
}

func (m *DivePlanModel) Insert(
	ownerID int,
	name string,
	notes string,
	isSoloDive bool,
	descentRate float64,
	ascentRate float64,
	sacRate float64,
	tankCount int,
	tankVolume float64,
	workingPressure int,
	diveFactor float64,
	fn2 float64,
	fhe float64,
	maxPPO2 float64,
	stops []DivePlanStopInput,
) (int, error) {
	stmt := `
        insert into dive_plans (
            owner_id, name, notes, is_solo_dive, descent_rate, ascent_rate,
            sac_rate, tank_count, tank_volume, working_pressure, dive_factor,
            fn2, fhe, max_ppo2
        ) values (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
        )
        returning id
    `

	if len(stops) == 0 {
		return 0, fmt.Errorf("no stops provided, cannot save dive plan")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to start db transaction: %w", err)
	}
	defer tx.Rollback()

	result := tx.QueryRowContext(
		ctx,
		stmt,
		ownerID,
		name,
		notes,
		isSoloDive,
		descentRate,
		ascentRate,
		sacRate,
		tankCount,
		tankVolume,
		workingPressure,
		diveFactor,
		fn2,
		fhe,
		maxPPO2,
	)

	var id int
	err = result.Scan(&id)
	if err != nil {
		return 0, err
	}

	err = insertDivePlanStops(ctx, tx, false, id, stops)
	if err != nil {
		return 0, fmt.Errorf("failed to insert plan stops: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("failed to commit db transaction: %w", err)
	}

	return id, nil
}

func (m *DivePlanModel) Update(
	id int,
	ownerID int,
	name string,
	notes string,
	isSoloDive bool,
	descentRate float64,
	ascentRate float64,
	sacRate float64,
	tankCount int,
	tankVolume float64,
	workingPressure int,
	diveFactor float64,
	fn2 float64,
	fhe float64,
	maxPPO2 float64,
	stops []DivePlanStopInput,
) error {
	stmt := `
        update dive_plans
           set version = version + 1, updated_at = now(), name = $3, notes = $4,
               is_solo_dive = $5, descent_rate = $6, ascent_rate = $7,
               sac_rate = $8, tank_count = $9, tank_volume = $10,
               working_pressure = $11, dive_factor = $12, fn2 = $13, fhe = $14,
               max_ppo2 = $15
         where id = $1
           and owner_id = $2
    `

	if len(stops) == 0 {
		return fmt.Errorf("no stops provided, cannot update dive plan")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start db transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		stmt,
		id,
		ownerID,
		name,
		notes,
		isSoloDive,
		descentRate,
		ascentRate,
		sacRate,
		tankCount,
		tankVolume,
		workingPressure,
		diveFactor,
		fn2,
		fhe,
		maxPPO2,
	)

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

	err = insertDivePlanStops(ctx, tx, true, id, stops)
	if err != nil {
		return fmt.Errorf("failed to insert plan stops: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit db transaction: %w", err)
	}

	return nil
}

func (m *DivePlanModel) GetOneByID(id, ownerID int) (DivePlan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var totalRecords int
	var divePlan DivePlan
	row := m.DB.QueryRowContext(ctx, divePlanSelectQuery, ownerID, id)
	err := divePlanFromDBRow(row, &totalRecords, &divePlan)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return DivePlan{}, ErrNoRecord
		} else {
			return DivePlan{}, err
		}
	}

	return divePlan, nil
}

func (m *DivePlanModel) List(
	diverID int,
	filters Pager,
	sort []SortDivePlan,
) ([]DivePlan, PageData, error) {
	limit := filters.limit()
	offset := filters.offset()
	order := buildOrderByClause(sort, SortDivePlanIDAsc)
	stmt := fmt.Sprintf("%s %s limit $3 offset $4", divePlanSelectQuery, order)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, stmt, diverID, nil, limit, offset)
	if err != nil {
		return nil, PageData{}, err
	}
	defer rows.Close()

	var totalRecords int
	var divePlans []DivePlan
	for rows.Next() {
		var dp DivePlan
		err := divePlanFromDBRow(rows, &totalRecords, &dp)
		if err != nil {
			return nil, PageData{}, err
		}
		divePlans = append(divePlans, dp)
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

	return divePlans, paginationData, nil
}

func (m *DivePlanModel) Exists(id int) (bool, error) {
	return idExistsInTable(m.DB, id, "dive_plans", "id")
}
