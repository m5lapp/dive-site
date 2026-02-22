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

// buildOrderByClause generates the order by clause for a SQL statement using
// each of the given table cols in order. The defaultCol sortCol is applied at
// the end with the intention of providing always-consistent ordering, it is
// recommended to use a unique or primary key column for this parameter.
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

// Buddy sorting options.
type SortBuddy struct{ sortCol }

func (SortBuddy) isSort() {}

var (
	SortBuddyIDAsc  = SortBuddy{sortCol{column: "bu.id", direction: sortAsc}}
	SortBuddyIDDesc = SortBuddy{sortCol{column: "bu.id", direction: sortDesc}}

	SortBuddyNameAsc  = SortBuddy{sortCol{column: "bu.name", direction: sortAsc}}
	SortBuddyNameDesc = SortBuddy{sortCol{column: "bu.name", direction: sortDesc}}

	SortBuddyDivesWithAsc  = SortBuddy{sortCol{column: "dives_with", direction: sortAsc}}
	SortBuddyDivesWithDesc = SortBuddy{sortCol{column: "dives_with", direction: sortDesc}}

	SortBuddyFirstDiveWithAsc = SortBuddy{
		sortCol{column: "ds.first_dive_with", direction: sortAsc},
	}
	SortBuddyFirstDiveWithDesc = SortBuddy{
		sortCol{column: "ds.first_dive_with", direction: sortDesc},
	}

	SortBuddyDefault = []SortBuddy{SortBuddyNameAsc, SortBuddyIDAsc}
)

// Certification sorting options.
type SortCert struct{ sortCol }

func (SortCert) isSort() {}

var (
	SortCertIDAsc  = SortCert{sortCol{column: "ce.id", direction: sortAsc}}
	SortCertIDDesc = SortCert{sortCol{column: "ce.id", direction: sortDesc}}

	SortCertAgencyAsc  = SortCert{sortCol{column: "ag.common_name", direction: sortAsc}}
	SortCertAgencyDesc = SortCert{sortCol{column: "ag.common_name", direction: sortDesc}}

	SortCertNameAsc  = SortCert{sortCol{column: "ac.name", direction: sortAsc}}
	SortCertNameDesc = SortCert{sortCol{column: "ac.name", direction: sortDesc}}

	SortCertStartDateAsc  = SortCert{sortCol{column: "ce.start_date", direction: sortAsc}}
	SortCertStartDateDesc = SortCert{sortCol{column: "ce.start_date", direction: sortDesc}}

	SortCertDefault = []SortCert{
		SortCertStartDateDesc,
		SortCertAgencyAsc,
		SortCertNameAsc,
		SortCertIDAsc,
	}
)

// Dive sorting options.
type SortDive struct{ sortCol }

func (SortDive) isSort() {}

var (
	SortDiveIDAsc  = SortDive{sortCol{column: "dv.id", direction: sortAsc}}
	SortDiveIDDesc = SortDive{sortCol{column: "dv.id", direction: sortDesc}}

	SortDiveDateAsc  = SortDive{sortCol{column: "dv.date_time_in", direction: sortAsc}}
	SortDiveDateDesc = SortDive{sortCol{column: "dv.date_time_in", direction: sortDesc}}

	SortDiveDefault = []SortDive{SortDiveDateDesc, SortDiveIDAsc}
)

// DivePlan sorting options.
type SortDivePlan struct{ sortCol }

func (SortDivePlan) isSort() {}

var (
	SortDivePlanIDAsc  = SortDivePlan{sortCol{column: "dp.id", direction: sortAsc}}
	SortDivePlanIDDesc = SortDivePlan{sortCol{column: "dp.id", direction: sortDesc}}

	SortDivePlanCreatedAsc  = SortDivePlan{sortCol{column: "dp.created_at", direction: sortAsc}}
	SortDivePlanCreatedDesc = SortDivePlan{sortCol{column: "dp.created_at", direction: sortDesc}}

	SortDivePlanIsSoloDiveAsc = SortDivePlan{
		sortCol{column: "dp.is_solo_dive", direction: sortAsc},
	}
	SortDivePlanIsSoloDiveDesc = SortDivePlan{
		sortCol{column: "dp.is_solo_dive", direction: sortDesc},
	}

	SortDivePlanNameAsc  = SortDivePlan{sortCol{column: "dp.name", direction: sortAsc}}
	SortDivePlanNameDesc = SortDivePlan{sortCol{column: "dp.name", direction: sortDesc}}

	SortDivePlanDefault = []SortDivePlan{
		SortDivePlanNameAsc,
	}
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

// Operator sorting options.
type SortOperator struct{ sortCol }

func (SortOperator) isSort() {}

var (
	SortOperatorIDAsc  = SortOperator{sortCol{column: "op.id", direction: sortAsc}}
	SortOperatorIDDesc = SortOperator{sortCol{column: "op.id", direction: sortDesc}}

	SortOperatorNameAsc  = SortOperator{sortCol{column: "op.name", direction: sortAsc}}
	SortOperatorNameDesc = SortOperator{sortCol{column: "op.name", direction: sortDesc}}

	SortOperatorTypeAsc  = SortOperator{sortCol{column: "ot.name", direction: sortAsc}}
	SortOperatorTypeDesc = SortOperator{sortCol{column: "ot.name", direction: sortDesc}}

	SortOperatorStreetAsc  = SortOperator{sortCol{column: "op.street", direction: sortAsc}}
	SortOperatorStreetDesc = SortOperator{sortCol{column: "op.street", direction: sortDesc}}

	SortOperatorSuburbAsc  = SortOperator{sortCol{column: "op.suburb", direction: sortAsc}}
	SortOperatorSuburbDesc = SortOperator{sortCol{column: "op.suburb", direction: sortDesc}}

	SortOperatorStateAsc  = SortOperator{sortCol{column: "op.state", direction: sortAsc}}
	SortOperatorStateDesc = SortOperator{sortCol{column: "op.state", direction: sortDesc}}

	SortOperatorPostcodeAsc  = SortOperator{sortCol{column: "op.postcode", direction: sortAsc}}
	SortOperatorPostcodeDesc = SortOperator{sortCol{column: "op.postcode", direction: sortDesc}}

	SortOperatorCountryAsc  = SortOperator{sortCol{column: "co.name", direction: sortAsc}}
	SortOperatorCountryDesc = SortOperator{sortCol{column: "co.name", direction: sortDesc}}

	SortOperatorDefault = []SortOperator{
		SortOperatorCountryAsc,
		SortOperatorNameAsc,
		SortOperatorStateAsc,
		SortOperatorSuburbAsc,
		SortOperatorStreetAsc,
		SortOperatorTypeAsc,
		SortOperatorIDAsc,
	}
)

// Trip sorting options.
type SortTrip struct{ sortCol }

func (SortTrip) isSort() {}

var (
	SortTripIDAsc  = SortTrip{sortCol{column: "tr.id", direction: sortAsc}}
	SortTripIDDesc = SortTrip{sortCol{column: "tr.id", direction: sortDesc}}

	SortTripNameAsc  = SortTrip{sortCol{column: "tr.name", direction: sortAsc}}
	SortTripNameDesc = SortTrip{sortCol{column: "tr.name", direction: sortDesc}}

	SortTripStartDateAsc  = SortTrip{sortCol{column: "tr.start_date", direction: sortAsc}}
	SortTripStartDateDesc = SortTrip{sortCol{column: "tr.start_date", direction: sortDesc}}

	SortTripDefault = []SortTrip{SortTripStartDateDesc, SortTripNameAsc, SortTripIDAsc}
)
