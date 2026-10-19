package ownership

import (
	"testing"
	"time"
)

// The two records below are shaped on the live zones response of 20 October
// 2026, in which 11 of 1,245 zones omitted both `currentOwner` and
// `dateLastTaken` as keys — never sent as null.
const (
	takenZone = `{"id":608535,"name":"Taken","totalTakeovers":1180,
		"currentOwner":{"name":"Zinky","id":330825},"dateLastTaken":"2026-10-17T14:03:29+0000"}`
	untakenZone = `{"id":808912,"name":"Untaken","totalTakeovers":17}`
)

// TestAnAbsentCurrentOwnerStaysAbsent is #22 AC1.
func TestAnAbsentCurrentOwnerStaysAbsent(t *testing.T) {
	t.Parallel()

	zones, err := ParseZones([]byte(`[` + untakenZone + `,` + takenZone + `]`))
	if err != nil {
		t.Fatalf("a zone response omitting currentOwner was refused: %v, want it processed", err)
	}

	if got := zones[0].CurrentOwner; got != nil {
		t.Errorf("zone %d omits currentOwner and was carried with owner %+v, want it carried as absent", zones[0].ID, *got)
	}

	if got := zones[1].CurrentOwner; got == nil || *got != (Owner{ID: 330825, Name: "Zinky"}) {
		t.Errorf("zone %d carries currentOwner Zinky (330825) and was read as %+v", zones[1].ID, got)
	}
}

// TestAnAbsentDateLastTakenStaysAbsent is #22 AC2.
func TestAnAbsentDateLastTakenStaysAbsent(t *testing.T) {
	t.Parallel()

	zones, err := ParseZones([]byte(`[` + untakenZone + `,` + takenZone + `]`))
	if err != nil {
		t.Fatalf("a zone response omitting dateLastTaken was refused: %v, want it processed", err)
	}

	if got := zones[0].DateLastTaken; got != nil {
		t.Errorf("zone %d omits dateLastTaken and was carried with date %s, want it carried as absent", zones[0].ID, got)
	}

	want := time.Date(2026, 10, 17, 14, 3, 29, 0, time.UTC)
	if got := zones[1].DateLastTaken; got == nil || !got.Equal(want) {
		t.Errorf("zone %d carries dateLastTaken 2026-10-17T14:03:29+0000 and was read as %v", zones[1].ID, got)
	}
}

// TestEitherFieldMayBeAbsentAlone: FR-030's verification names either field,
// or both, so each is carried independently of the other.
func TestEitherFieldMayBeAbsentAlone(t *testing.T) {
	t.Parallel()

	zones, err := ParseZones([]byte(`[
		{"id":1,"currentOwner":{"name":"a","id":2}},
		{"id":3,"dateLastTaken":"2026-10-01T00:00:00+0000"}
	]`))
	if err != nil {
		t.Fatalf("processing: %v", err)
	}

	if zones[0].CurrentOwner == nil || zones[0].DateLastTaken != nil {
		t.Errorf("zone 1 carries an owner and no date, and was read as owner %v date %v", zones[0].CurrentOwner, zones[0].DateLastTaken)
	}

	if zones[1].CurrentOwner != nil || zones[1].DateLastTaken == nil {
		t.Errorf("zone 3 carries a date and no owner, and was read as owner %v date %v", zones[1].CurrentOwner, zones[1].DateLastTaken)
	}
}

// TestAZoneResponseItCannotReadIsRefused: absence is data, but a malformed
// record is not, and is never defaulted into one.
func TestAZoneResponseItCannotReadIsRefused(t *testing.T) {
	t.Parallel()

	for name, body := range map[string]string{
		"not an array":         `{"id":1}`,
		"a record with no id":  `[{"name":"x"}]`,
		"an owner with no id":  `[{"id":1,"currentOwner":{"name":"x"}}]`,
		"an undocumented date": `[{"id":1,"dateLastTaken":"17/10/2026"}]`,
	} {
		if _, err := ParseZones([]byte(body)); err == nil {
			t.Errorf("%s: processed without error, want it refused", name)
		}
	}
}
