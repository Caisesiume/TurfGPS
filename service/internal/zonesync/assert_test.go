package zonesync

import (
	"strings"
	"testing"
)

// TestTheFourAssertionsEachRefuseTheirOwnDefect walks the list in
// `Architecture.md § The sync write path`. Each case is a staged response that
// is fine in every respect but one.
func TestTheFourAssertionsEachRefuseTheirOwnDefect(t *testing.T) {
	t.Parallel()

	// A response that clears all four, used as the baseline every case below
	// spoils exactly one property of.
	clean := Staged{Rows: 1000, DistinctIDs: 1000, CurrentZoneRows: 990}

	if reasons := clean.Verify(0.9); len(reasons) != 0 {
		t.Fatalf("the baseline was refused for %v, so every case below would be measuring the wrong thing", reasons)
	}

	cases := map[string]struct {
		staged Staged
		names  string
	}{
		"a truncated response": {
			staged: Staged{Rows: 400, DistinctIDs: 400, CurrentZoneRows: 1000},
			names:  "floor",
		},
		"duplicate ids": {
			staged: Staged{Rows: 1000, DistinctIDs: 998, CurrentZoneRows: 990},
			names:  "distinct ids",
		},
		"a null id": {
			staged: Staged{Rows: 1000, DistinctIDs: 999, NullIDs: 1, CurrentZoneRows: 990},
			names:  "no id",
		},
		"a coordinate out of range": {
			staged: Staged{Rows: 1000, DistinctIDs: 1000, OutOfRange: 3, CurrentZoneRows: 990},
			names:  "outside the permitted range",
		},
		"a dateCreated after the run started": {
			staged: Staged{Rows: 1000, DistinctIDs: 1000, FutureCreated: 2, CurrentZoneRows: 990},
			names:  "later than this run's start",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			reasons := tc.staged.Verify(0.9)
			if len(reasons) == 0 {
				t.Fatalf("%s was accepted for merge, want it refused", name)
			}

			if !strings.Contains(strings.Join(reasons, " | "), tc.names) {
				t.Errorf("the refusal reads %v, want it to name what was wrong (%q)", reasons, tc.names)
			}
		})
	}
}

// TestAnEmptyResponseIsRefusedAgainstAnEmptyTable is the case the floor alone
// cannot catch, and the reason the emptiness check exists beside it.
//
// Against an empty `zone` the floor is zero, so every count clears it — and the
// first run of a fresh store would record `ok` for a response that carried
// nothing at all, which is the strongest possible truncation.
func TestAnEmptyResponseIsRefusedAgainstAnEmptyTable(t *testing.T) {
	t.Parallel()

	empty := Staged{Rows: 0, DistinctIDs: 0, CurrentZoneRows: 0}

	if reasons := empty.Verify(0.9); len(reasons) == 0 {
		t.Error("an empty response against an empty table was accepted for merge, want it refused")
	}
}

// TestTheFirstRealResponseIsAcceptedAgainstAnEmptyTable is the other side of the
// same boundary, and it is worth pinning: a floor computed from a table holding
// nothing must not refuse the run that fills it.
func TestTheFirstRealResponseIsAcceptedAgainstAnEmptyTable(t *testing.T) {
	t.Parallel()

	first := Staged{Rows: 154, DistinctIDs: 154, CurrentZoneRows: 0}

	if reasons := first.Verify(0.9); len(reasons) != 0 {
		t.Errorf("the first response into an empty store was refused for %v, want it merged", reasons)
	}
}

// TestEveryDefectIsReportedAndNotJustTheFirst binds why Verify returns a slice.
// All four are computed by one query, and a run refused should record what was
// wrong with the response rather than that something was.
func TestEveryDefectIsReportedAndNotJustTheFirst(t *testing.T) {
	t.Parallel()

	bad := Staged{Rows: 10, DistinctIDs: 8, NullIDs: 1, OutOfRange: 2, FutureCreated: 1, CurrentZoneRows: 1000}

	if reasons := bad.Verify(0.9); len(reasons) < 4 {
		t.Errorf("a response failing every assertion produced %d reasons (%v), want one per defect", len(reasons), reasons)
	}
}
