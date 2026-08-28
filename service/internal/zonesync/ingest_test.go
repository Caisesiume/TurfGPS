package zonesync

import (
	"strings"
	"testing"
	"time"
)

// TestTheOrdinaryRecordIsTheOneWithoutTheOptionalFields binds the mapping's
// least obvious property: a zone carrying no `type`, no `region.country` and no
// `region.area` is the ORDINARY case and not an anomaly.
//
// `Architecture.md § Retrieving zones` measures `type` on fewer than one zone in
// six and `region.country` missing on thousands, and records that a schema
// treating either as required rejects perfectly ordinary records. A parser that
// refused them would make the sync fail on most of the corpus.
func TestTheOrdinaryRecordIsTheOneWithoutTheOptionalFields(t *testing.T) {
	t.Parallel()

	const body = `[{"id":42,"name":"Untyped","latitude":-33.87,"longitude":151.21,
	  "dateCreated":"2019-01-02T03:04:05Z","totalTakeovers":318,"takeoverPoints":75,
	  "pointsPerHour":2,"region":{"id":7,"name":"Australia"}}]`

	zones, err := ParseAllZones([]byte(body))
	if err != nil {
		t.Fatalf("a record with no type, country or area was refused: %v", err)
	}

	if len(zones) != 1 {
		t.Fatalf("parsed %d zones, want 1", len(zones))
	}

	z := zones[0]

	if z.TypeID != nil {
		t.Errorf("type_id = %d, want it null: an absent type is the ordinary case and never an anomaly", *z.TypeID)
	}

	if z.RegionCountry != nil {
		t.Errorf("region_country = %q, want it null", *z.RegionCountry)
	}

	if z.AreaID != nil || z.AreaName != nil {
		t.Errorf("area_id = %v and area_name = %v, want both null", z.AreaID, z.AreaName)
	}

	if z.RegionName != "Australia" {
		t.Errorf("region_name = %q, want %q — for a country Turf has not subdivided this column holds the country's own name", z.RegionName, "Australia")
	}
}

// TestTheFullRecordMapsOntoEveryStagingColumn is the mapping in the other
// direction: every field the response carries reaches the column
// `migrations/README.md § The zone ingest field mapping` names for it.
func TestTheFullRecordMapsOntoEveryStagingColumn(t *testing.T) {
	t.Parallel()

	const body = `[{"id":660960,"name":"Resited","latitude":57.708870,"longitude":11.974560,
	  "dateCreated":"2025-10-30T00:00:00Z","totalTakeovers":10,"takeoverPoints":85,
	  "pointsPerHour":4,"type":{"id":3,"name":"water"},
	  "region":{"id":142,"name":"Göteborg","country":"SE",
	            "area":{"id":9001,"name":"Centrum"}}}]`

	zones, err := ParseAllZones([]byte(body))
	if err != nil {
		t.Fatalf("parsing a complete record: %v", err)
	}

	got := zones[0]

	want := Zone{
		ID:             660960,
		Name:           "Resited",
		Latitude:       57.708870,
		Longitude:      11.974560,
		DateCreated:    time.Date(2025, 10, 30, 0, 0, 0, 0, time.UTC),
		TotalTakeovers: 10,
		TakeoverPoints: 85,
		PointsPerHour:  4,
		RegionID:       142,
		RegionName:     "Göteborg",
	}

	if got.ID != want.ID || got.Name != want.Name || got.Latitude != want.Latitude ||
		got.Longitude != want.Longitude || !got.DateCreated.Equal(want.DateCreated) ||
		got.TotalTakeovers != want.TotalTakeovers || got.TakeoverPoints != want.TakeoverPoints ||
		got.PointsPerHour != want.PointsPerHour || got.RegionID != want.RegionID ||
		got.RegionName != want.RegionName {
		t.Errorf("the scalar columns mapped to %+v, want %+v", got, want)
	}

	if got.TypeID == nil || *got.TypeID != 3 {
		t.Errorf("type_id = %v, want 3 — the id is stored and type.name is dropped", got.TypeID)
	}

	if got.RegionCountry == nil || *got.RegionCountry != "SE" {
		t.Errorf("region_country = %v, want SE", got.RegionCountry)
	}

	if got.AreaID == nil || *got.AreaID != 9001 || got.AreaName == nil || *got.AreaName != "Centrum" {
		t.Errorf("area_id = %v and area_name = %v, want 9001 and Centrum", got.AreaID, got.AreaName)
	}
}

// TestACoordinateIsNeverInvertedByTheMapping is the one mapping error with a
// silent consequence: latitude and longitude are two scalars in the response,
// they are stored as two scalars, and `zone.geom` is generated from them in DDL.
// Swapping them here would put every zone somewhere else with no error anywhere.
func TestACoordinateIsNeverInvertedByTheMapping(t *testing.T) {
	t.Parallel()

	const body = `[{"id":1,"name":"Göteborg","latitude":57.7,"longitude":11.97,
	  "dateCreated":"2020-01-01T00:00:00Z","totalTakeovers":1,"takeoverPoints":50,
	  "pointsPerHour":1,"region":{"id":1,"name":"Sverige"}}]`

	zones, err := ParseAllZones([]byte(body))
	if err != nil {
		t.Fatalf("parsing: %v", err)
	}

	if zones[0].Latitude != 57.7 {
		t.Errorf("latitude = %v, want 57.7 — the response's own latitude, not its longitude", zones[0].Latitude)
	}

	if zones[0].Longitude != 11.97 {
		t.Errorf("longitude = %v, want 11.97 — the response's own longitude, not its latitude", zones[0].Longitude)
	}
}

// TestAMissingRequiredFieldIsRefusedRatherThanDefaulted binds why every wire
// field is a pointer. Zero is a legitimate value for several of these, so a
// value type could not tell an omitted field from one sent as zero — and the
// omitted one would be merged as a real zero into a NOT NULL column.
func TestAMissingRequiredFieldIsRefusedRatherThanDefaulted(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"no id":              `[{"name":"x","latitude":1,"longitude":2,"dateCreated":"2020-01-01T00:00:00Z","totalTakeovers":0,"takeoverPoints":0,"pointsPerHour":0,"region":{"id":1,"name":"r"}}]`,
		"no coordinate":      `[{"id":1,"name":"x","dateCreated":"2020-01-01T00:00:00Z","totalTakeovers":0,"takeoverPoints":0,"pointsPerHour":0,"region":{"id":1,"name":"r"}}]`,
		"no dateCreated":     `[{"id":1,"name":"x","latitude":1,"longitude":2,"totalTakeovers":0,"takeoverPoints":0,"pointsPerHour":0,"region":{"id":1,"name":"r"}}]`,
		"no totalTakeovers":  `[{"id":1,"name":"x","latitude":1,"longitude":2,"dateCreated":"2020-01-01T00:00:00Z","takeoverPoints":0,"pointsPerHour":0,"region":{"id":1,"name":"r"}}]`,
		"no region":          `[{"id":1,"name":"x","latitude":1,"longitude":2,"dateCreated":"2020-01-01T00:00:00Z","totalTakeovers":0,"takeoverPoints":0,"pointsPerHour":0}]`,
		"no region name":     `[{"id":1,"name":"x","latitude":1,"longitude":2,"dateCreated":"2020-01-01T00:00:00Z","totalTakeovers":0,"takeoverPoints":0,"pointsPerHour":0,"region":{"id":1}}]`,
		"an unreadable date": `[{"id":1,"name":"x","latitude":1,"longitude":2,"dateCreated":"last Tuesday","totalTakeovers":0,"takeoverPoints":0,"pointsPerHour":0,"region":{"id":1,"name":"r"}}]`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if zones, err := ParseAllZones([]byte(body)); err == nil {
				t.Errorf("the response was accepted as %+v, want it refused: a record missing a field measured on every zone is a response this system does not understand, not a zone with a default", zones)
			}
		})
	}
}

// TestAResponseThatIsNotAZoneArrayIsRefused covers the shape rather than the
// records. It is the failure that reaches `http_error` as "a body that would not
// parse", and it must not be read as an empty corpus.
func TestAResponseThatIsNotAZoneArrayIsRefused(t *testing.T) {
	t.Parallel()

	for name, body := range map[string]string{
		"an object":        `{"zones":[]}`,
		"an error payload": `{"status":"error"}`,
		"truncated":        `[{"id":1,`,
		"trailing content": `[] []`,
		"empty":            ``,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := ParseAllZones([]byte(body)); err == nil {
				t.Error("the response was accepted, want it refused")
			}
		})
	}
}

// TestAnEmptyArrayIsAnEmptyResponseAndNotAnError draws the line the test above
// stops at: a well-formed response carrying no zones parses, and it is the
// staging assertions that refuse to merge it. Conflating the two would put a
// truncation verdict in the parser, where the current table's size is not known.
func TestAnEmptyArrayIsAnEmptyResponseAndNotAnError(t *testing.T) {
	t.Parallel()

	zones, err := ParseAllZones([]byte(`[]`))
	if err != nil {
		t.Fatalf("an empty array was refused by the parser: %v", err)
	}

	if len(zones) != 0 {
		t.Errorf("parsed %d zones from an empty array", len(zones))
	}

	if reasons := (Staged{Rows: 0, CurrentZoneRows: 0}).Verify(0.9); len(reasons) == 0 {
		t.Error("an empty response was accepted for merge, want it refused by the staging assertions")
	}
}

// TestADateWithNoZoneIsReadAsUTC binds the one place the host could leak into
// the data. Reading a zoneless timestamp in the host's local zone would make the
// stored instant depend on where the service runs, which shows up as a shifted
// takeover-rate denominator rather than as an error.
func TestADateWithNoZoneIsReadAsUTC(t *testing.T) {
	t.Parallel()

	const body = `[{"id":1,"name":"x","latitude":1,"longitude":2,
	  "dateCreated":"2024-04-28T10:00:00","totalTakeovers":0,"takeoverPoints":0,
	  "pointsPerHour":0,"region":{"id":1,"name":"r"}}]`

	zones, err := ParseAllZones([]byte(body))
	if err != nil {
		t.Fatalf("parsing: %v", err)
	}

	want := time.Date(2024, 4, 28, 10, 0, 0, 0, time.UTC)
	if !zones[0].DateCreated.Equal(want) {
		t.Errorf("dateCreated = %s, want %s", zones[0].DateCreated, want)
	}
}

// TestTheRefusalNamesTheRecord keeps the failure diagnosable. A sync that
// refuses a whole corpus because one of its records is malformed has to say
// which one.
func TestTheRefusalNamesTheRecord(t *testing.T) {
	t.Parallel()

	const body = `[{"id":1,"name":"x","latitude":1,"longitude":2,"dateCreated":"2020-01-01T00:00:00Z","totalTakeovers":0,"takeoverPoints":0,"pointsPerHour":0,"region":{"id":1,"name":"r"}},
	              {"id":2,"name":"y","latitude":1,"longitude":2,"dateCreated":"2020-01-01T00:00:00Z","totalTakeovers":0,"takeoverPoints":0,"pointsPerHour":0}]`

	_, err := ParseAllZones([]byte(body))
	if err == nil {
		t.Fatal("the response was accepted, want it refused")
	}

	if !strings.Contains(err.Error(), "record 1") {
		t.Errorf("the refusal reads %q, want it to name the offending record", err)
	}
}
