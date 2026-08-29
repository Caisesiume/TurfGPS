package syncstore

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Caisesiume/TurfGPS/service/internal/zonesync"
)

// ---------------------------------------------------------------------------
// THIS FILE IS DATABASE-FREE AND LEAVES THE SQL HALF OF THIS PACKAGE'S
// NEVER-EXECUTED MARKER STANDING. It falsified the other half. Read that marker
// at the top of syncstore.go before adding to this file, and read which of its
// two claims this file is on each side of:
//
//   - STANDING, and this file does not touch it: not one statement in
//     syncstore.go has been sent to a PostgreSQL server. Nothing here sends one,
//     opens a pool, or fakes one, so what PostgreSQL does with those statements
//     is exactly as unmeasured as it was before this file existed.
//   - FALSIFIED, by this file's existence: the marker also said the package
//     carried no tests and that there was no test file in this directory. Those
//     sentences were narrowed away when this file landed, rather than left to be
//     read as the standing half.
//
// The distinction is the file's whole licence to be here. What it measures is a
// question decided inside this process, before any statement is sent — which
// value is bound to which column name — and that question has an answer whether
// or not a database exists. What it must never do is let a green run here be
// read as evidence about the half above it that is still standing.
//
// THE BINDING IS THE ONE ERROR NO GUARD OVER THE WRITTEN ZONES CAN SEE, and it
// is worth being exact about why, because four of them stand over the
// coordinate and a crossed pair walks past all four.
//
//   - The two range constraints on `zone` admit it. A Swedish latitude is a
//     valid longitude and the reverse, so the pair is in range whichever column
//     receives which value, and nothing is rejected.
//   - The two checks that compare `geom` against the scalars cannot fail by
//     construction. `geom` is GENERATED from `latitude` and `longitude` by the
//     DDL, so a crossing moves both sides of that comparison together: the
//     point is built from whatever landed in the columns and agrees with them
//     exactly as it did before, while every zone in the country sits somewhere
//     else.
//
// The compiler cannot object either — the two fields are both float64, so the
// crossing is a legal program. Eleven of the fourteen columns share a type with
// at least one other, so the same silence covers the id/total_takeovers,
// name/region_name, takeover_points/points_per_hour/region_id and
// region_country/area_name groups; a crossed area name is merely wrong, but it
// is wrong just as quietly.
//
// SO THE PINNED SET IS WRITTEN OUT A SECOND TIME, BY HAND, AND NOT DERIVED.
// A test that read the column name and the accessor from the same slice would
// agree with any crossing that slice happened to contain. pinnedColumns states
// the column names, in the migration's declared order, against literal values;
// probeZone states the same literals against Go field names. The only thing
// joining the two statements is zoneColumns, which is the thing under test.
// ---------------------------------------------------------------------------

// pinnedColumn is one expected column: its name, its position by virtue of this
// slice's order, and the value the probe zone supplies for it.
type pinnedColumn struct {
	name  string
	value any
}

// pinnedColumns is zone_incoming restated in the migration's declared order.
//
// THE ORDER IS PART OF THE ASSERTION. CopyFrom is positional against the column
// list derived from zoneColumns, so rearranging that slice rearranges what COPY
// is handed; the order is therefore pinned here rather than left to whoever
// edits it next.
//
// EVERY VALUE DIFFERS FROM EVERY OTHER, which is the property that makes a
// crossing visible at all. Two same-typed columns carrying the same probe value
// would swap undetected, so this is a constraint on the literals and not a
// stylistic choice. It is asserted by TestEveryPinnedValueDiffersFromEveryOther
// rather than left to whoever edits this slice next: a property that only the
// comments hold is a property that degrades on the one edit nobody re-reads the
// comments for, and it degrades silently — every test in this file still passes
// while the crossing it exists to catch walks through.
func pinnedColumns() []pinnedColumn {
	return []pinnedColumn{
		{"id", int32(1001)},
		{"name", "the zone's own name"},
		{"latitude", 57.7},
		{"longitude", 11.97},
		{"date_created", time.Date(2020, 3, 4, 5, 6, 7, 0, time.UTC)},
		{"total_takeovers", int32(2002)},
		{"takeover_points", int16(301)},
		{"points_per_hour", int16(302)},
		{"type_id", ptrTo(int16(303))},
		{"region_id", int16(304)},
		{"region_name", "the region's own name"},
		{"region_country", ptrTo("SE")},
		{"area_id", ptrTo(int32(4004))},
		{"area_name", ptrTo("the area's own name")},
	}
}

// probeZone carries the same fourteen literals, stated against Go field names.
//
// The coordinate is Göteborg's, and is a real pair on purpose: 11.97 is a valid
// latitude and 57.7 a valid longitude, so a crossed pair is not out of range and
// raises nothing anywhere. It lands the zone at 11.97°N 57.7°E — the Arabian
// Sea, some 5,600 km away — which is the point: the crossing is undetectable by
// range and glaring by position, and no constraint in the schema looks at
// position.
func probeZone() zonesync.Zone {
	return zonesync.Zone{
		ID:             1001,
		Name:           "the zone's own name",
		Latitude:       57.7,
		Longitude:      11.97,
		DateCreated:    time.Date(2020, 3, 4, 5, 6, 7, 0, time.UTC),
		TotalTakeovers: 2002,
		TakeoverPoints: 301,
		PointsPerHour:  302,
		TypeID:         ptrTo(int16(303)),
		RegionID:       304,
		RegionName:     "the region's own name",
		RegionCountry:  ptrTo("SE"),
		AreaID:         ptrTo(int32(4004)),
		AreaName:       ptrTo("the area's own name"),
	}
}

func ptrTo[T any](v T) *T { return &v }

// show prints a value the way a failure message needs it. Four of the columns
// are nullable and arrive as pointers, and the default verb on a pointer prints
// an address, which tells a reader nothing about which field was read.
func show(v any) string {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return "nil"
		}

		return fmt.Sprintf("&%#v", rv.Elem().Interface())
	}

	return fmt.Sprintf("%#v", v)
}

// TestEveryPinnedValueDiffersFromEveryOther asserts the precondition every
// other test in this file rests on, and which pinnedColumns' doc had only
// declared.
//
// WITHOUT IT THE NET DEGRADES SILENTLY, WHICH IS THE ONLY REASON IT IS A TEST
// AND NOT A COMMENT. Give two same-typed columns the same probe value — the
// obvious thing to do when adding a column and reaching for a literal — and
// TestEveryColumnIsLoadedFromItsOwnField still passes, because a crossing
// between those two loads each of them with the value the other was pinned to
// and both comparisons succeed. Nothing goes red. The pair simply stops being
// covered, and the file goes on reporting that it covers all fourteen.
//
// PINNING IS ENOUGH, AND THE PROBE DOES NOT NEED ITS OWN CHECK. probeZone
// carries the same fourteen literals, and the test below compares what each
// accessor loads out of the probe against the pinned value — so two probe
// fields sharing a value while their pinned values differ is already a failure
// there. Distinctness over the pinned set therefore carries the probe with it,
// and a second loop over probeZone would assert the same property twice.
//
// The comparison is reflect.DeepEqual, the same one the binding test uses, so
// this measures indistinguishability by the standard actually applied. That
// matters for the four nullable columns: DeepEqual follows pointers, so two
// distinct *int16 addresses holding the same number count as the same value
// here, which is right — the address is not what a crossing would swap.
func TestEveryPinnedValueDiffersFromEveryOther(t *testing.T) {
	t.Parallel()

	pinned := pinnedColumns()

	for i := range pinned {
		for j := i + 1; j < len(pinned); j++ {
			if !reflect.DeepEqual(pinned[i].value, pinned[j].value) {
				continue
			}

			t.Errorf("the %q and %q columns are both pinned to %s — a crossing between them loads each with the value the other expects, so every assertion in this file passes and neither column is covered any more; give them different values",
				pinned[i].name, pinned[j].name, show(pinned[i].value))
		}
	}
}

// TestEveryColumnIsLoadedFromItsOwnField is the assertion the file header
// argues for: every entry, and not only the two the geometry is built from.
func TestEveryColumnIsLoadedFromItsOwnField(t *testing.T) {
	t.Parallel()

	probe := probeZone()
	pinned := pinnedColumns()

	if len(zoneColumns) != len(pinned) {
		t.Fatalf("zoneColumns has %d entries and zone_incoming is pinned here at %d — a column added, removed or duplicated in the write path changes what COPY is handed, and it moves migrations/0001_zone_store.sql and this list with it",
			len(zoneColumns), len(pinned))
	}

	for i, want := range pinned {
		got := zoneColumns[i]

		if got.name != want.name {
			t.Errorf("zoneColumns[%d] is the %q column, want %q — CopyFrom is positional against the list derived from this slice, so reordering it feeds every value to the wrong column from that point on",
				i, got.name, want.name)

			continue
		}

		if loaded := got.value(probe); !reflect.DeepEqual(loaded, want.value) {
			t.Errorf("the %q column is loaded with %s, want %s — its accessor reads a different field of zonesync.Zone than its name says, which the compiler accepts whenever the two fields share a type and which no check over zone.geom can detect, geom being generated from these same columns",
				got.name, show(loaded), show(want.value))
		}
	}
}

// TestEveryFieldOfAZoneReachesAColumn covers the one direction the pinning
// above cannot: a field added to the ingest type and to no column here is
// decoded from the response, carried the whole way through the sync, and
// silently never written. The merge would report every such row unchanged.
func TestEveryFieldOfAZoneReachesAColumn(t *testing.T) {
	t.Parallel()

	if fields := reflect.TypeOf(zonesync.Zone{}).NumField(); fields != len(zoneColumns) {
		t.Errorf("zonesync.Zone has %d fields and zoneColumns writes %d columns — a field the sync decodes but never writes reaches the database in no column at all",
			fields, len(zoneColumns))
	}
}

// TestTheMergeStatementIsExactlyAsPinned pins the statement buildMergeSQL
// derives from zoneColumns.
//
// IT IS A SEPARATE QUESTION FROM THE BINDING ABOVE, and neither test subsumes
// the other. Crossing two accessors leaves this statement byte-identical — no
// name and no position moves — so the pinned text reports green over the
// coordinate defect. Reordering two entries leaves every accessor bound to its
// own field and changes this text. The pair is what covers both.
func TestTheMergeStatementIsExactlyAsPinned(t *testing.T) {
	t.Parallel()

	got := buildMergeSQL()
	if got == wantMergeSQL {
		return
	}

	gotLines, wantLines := strings.Split(got, "\n"), strings.Split(wantMergeSQL, "\n")

	var differing int

	for i := 0; i < len(gotLines) || i < len(wantLines); i++ {
		g, w := beyondTheStatement, beyondTheStatement

		if i < len(gotLines) {
			g = gotLines[i]
		}

		if i < len(wantLines) {
			w = wantLines[i]
		}

		if g == w {
			continue
		}

		differing++

		t.Errorf("the merge statement differs from the pinned text at line %d:\n got: %q\nwant: %q",
			i+1, g, w)
	}

	if differing == 0 {
		t.Errorf("the merge statement differs from the pinned text, but every line of it agrees and the two have the same length — a line of one of them is the literal %q",
			beyondTheStatement)
	}

	// Reported once, after every differing line, and the two halves do
	// different work. The count says how far the change reaches: one line is a
	// column renamed, thirty is the set or its order rebuilt, and a reader who
	// was shown only the first difference cannot tell those apart. The
	// statement itself is what makes the repair one pass — a break that is
	// legitimate is repaired by replacing wantMergeSQL with the text below,
	// where before it was repaired by rebuilding the constant a line at a time
	// and rerunning to find the next difference.
	//
	// Printed raw rather than quoted, because it is meant to be pasted between
	// the backticks of wantMergeSQL, and %q would have to be unescaped first.
	t.Errorf("%d line(s) differ. The statement is built from zoneColumns, so this means the ingest column set, its order, or the columns the merge assigns and tests for change have moved — which is a schema change and moves migrations/0001_zone_store.sql with it. Read the derived statement against that migration before pinning it; if it is right, this is the text to pin:\n%s",
		differing, got)
}

// beyondTheStatement stands in for a line one statement has and the other does
// not, so a length difference is reported as a differing line like any other
// rather than falling out of the comparison.
const beyondTheStatement = "<the statement ends here>"

// wantMergeSQL is the statement as it stands, captured from buildMergeSQL and
// read against migrations/0001_zone_store.sql line by line before being pinned.
const wantMergeSQL = `
INSERT INTO zone (id,
                  name,
                  latitude,
                  longitude,
                  date_created,
                  total_takeovers,
                  takeover_points,
                  points_per_hour,
                  type_id,
                  region_id,
                  region_name,
                  region_country,
                  area_id,
                  area_name,
                  first_seen_at, last_changed_at)
SELECT i.id,
       i.name,
       i.latitude,
       i.longitude,
       i.date_created,
       i.total_takeovers,
       i.takeover_points,
       i.points_per_hour,
       i.type_id,
       i.region_id,
       i.region_name,
       i.region_country,
       i.area_id,
       i.area_name,
       $1, $1
FROM   zone_incoming i
ON CONFLICT (id) DO UPDATE SET
       name            = excluded.name,
       latitude        = excluded.latitude,
       longitude       = excluded.longitude,
       date_created    = excluded.date_created,
       total_takeovers = excluded.total_takeovers,
       takeover_points = excluded.takeover_points,
       points_per_hour = excluded.points_per_hour,
       type_id         = excluded.type_id,
       region_id       = excluded.region_id,
       region_name     = excluded.region_name,
       region_country  = excluded.region_country,
       area_id         = excluded.area_id,
       area_name       = excluded.area_name,
       last_changed_at = excluded.last_changed_at
WHERE  zone.name            IS DISTINCT FROM excluded.name
   OR  zone.latitude        IS DISTINCT FROM excluded.latitude
   OR  zone.longitude       IS DISTINCT FROM excluded.longitude
   OR  zone.date_created    IS DISTINCT FROM excluded.date_created
   OR  zone.total_takeovers IS DISTINCT FROM excluded.total_takeovers
   OR  zone.takeover_points IS DISTINCT FROM excluded.takeover_points
   OR  zone.points_per_hour IS DISTINCT FROM excluded.points_per_hour
   OR  zone.type_id         IS DISTINCT FROM excluded.type_id
   OR  zone.region_id       IS DISTINCT FROM excluded.region_id
   OR  zone.region_name     IS DISTINCT FROM excluded.region_name
   OR  zone.region_country  IS DISTINCT FROM excluded.region_country
   OR  zone.area_id         IS DISTINCT FROM excluded.area_id
   OR  zone.area_name       IS DISTINCT FROM excluded.area_name`
