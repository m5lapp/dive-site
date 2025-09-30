package models

import (
	"fmt"
	"strings"
)

type sortDirection bool

const (
	sortAsc  sortDirection = true
	sortDesc sortDirection = false
)

func (sd sortDirection) String() string {
	switch sd {
	case sortDesc:
		return "desc"
	default:
		return "asc"
	}
}

type sortCol struct {
	column    string
	direction sortDirection
}

func (sc sortCol) String() string {
	return fmt.Sprintf("%s %s", sc.column, sc.direction.String())
}

type Sorter interface {
	fmt.Stringer
	// This unexported function prevents external implementations.
	isSort()
}

// buildOrderByClause generates the order by clause for a SQL statement using each of
// the given table cols in order. The defaultCol sortCol is applied at the end
// with the intention always-consistent ordering, it is recommended to use a
// unique or primary key column for this parameter.
func buildOrderByClause[T Sorter](cols []T, defaultCol T) string {
	var clause strings.Builder
	clause.WriteString("order by ")

	for i, col := range cols {
		if i > 0 {
			clause.WriteString(", ")
		}

		clause.WriteString(col.String())
	}

	if len(cols) > 0 {
		clause.WriteString(", ")
	}
	clause.WriteString(defaultCol.String())

	return clause.String()
}

// Dive sorting options.
type SortDive struct{ sortCol }

func (SortDive) isSort() {}

var (
	SortDiveIDAsc  = SortDive{sortCol{column: "dv.id", direction: sortAsc}}
	SortDiveIDDesc = SortDive{sortCol{column: "dv.id", direction: sortDesc}}

	SortDiveDateAsc  = SortDive{sortCol{column: "dv.date_time_in", direction: sortAsc}}
	SortDiveDateDesc = SortDive{sortCol{column: "dv.date_time_in", direction: sortDesc}}

	SortDiveDefault = []SortDive{SortDiveDateDesc}
)

// DiveSite sorting options.
type SortDiveSite struct{ sortCol }

func (SortDiveSite) isSort() {}

var (
	SortDiveSiteIDAsc  = SortDiveSite{sortCol{column: "ds.id", direction: sortAsc}}
	SortDiveSiteIDDesc = SortDiveSite{sortCol{column: "ds.id", direction: sortDesc}}

	SortDiveSiteCountryAsc  = SortDiveSite{sortCol{column: "co.name", direction: sortAsc}}
	SortDiveSiteCountryDesc = SortDiveSite{sortCol{column: "co.name", direction: sortDesc}}

	SortDiveSiteLocationAsc  = SortDiveSite{sortCol{column: "ds.location", direction: sortAsc}}
	SortDiveSiteLocationDesc = SortDiveSite{sortCol{column: "ds.location", direction: sortDesc}}

	SortDiveSiteNameAsc  = SortDiveSite{sortCol{column: "ds.name", direction: sortAsc}}
	SortDiveSiteNameDesc = SortDiveSite{sortCol{column: "ds.name", direction: sortDesc}}

	SortDiveSiteRegionAsc  = SortDiveSite{sortCol{column: "ds.region", direction: sortAsc}}
	SortDiveSiteRegionDesc = SortDiveSite{sortCol{column: "ds.region", direction: sortDesc}}

	SortDiveSiteDefault = []SortDiveSite{
		SortDiveSiteCountryAsc,
		SortDiveSiteLocationAsc,
		SortDiveSiteRegionAsc,
		SortDiveSiteNameAsc,
	}
)
