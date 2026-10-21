package ownership

import (
	"context"
	"testing"
)

// The end-to-end request counts for #24 are in internal/turf/session_test.go,
// against the real adapter; these cover the scan.

func regionsOnce(body string) FetchRegionsFunc {
	return func(context.Context) ([]byte, error) { return []byte(body), nil }
}

// TestALordOfSeveralRegionsIsHeld: one player may lead several regions — the
// live response of 20 October 2026 had one lord of ten — and that is still one
// value, held.
func TestALordOfSeveralRegionsIsHeld(t *testing.T) {
	t.Parallel()

	held, err := RegionLord(t.Context(), regionsOnce(`[
		{"id":1,"regionLord":{"name":"a","id":7}},
		{"id":2,"regionLord":{"name":"a","id":7}},
		{"id":3}
	]`), 7)
	if err != nil || !held {
		t.Errorf("a lord of two regions resolved held=%v err=%v, want held", held, err)
	}
}

func TestARegionResponseItCannotScanIsRefused(t *testing.T) {
	t.Parallel()

	for name, body := range map[string]string{
		"an empty list":     `[]`,
		"null":              `null`,
		"not a list":        `{"id":1}`,
		"a lord with no id": `[{"id":1,"regionLord":{"name":"a"}}]`,
	} {
		if _, err := RegionLord(t.Context(), regionsOnce(body), 7); err == nil {
			t.Errorf("%s: resolved without error, want it refused", name)
		}
	}
}

// TestANonPositiveIDIsNeverCompared is #234 S1: a user id of 0 and a lord id
// of 0 would compare equal and resolve held. Each side is refused on its own.
func TestANonPositiveIDIsNeverCompared(t *testing.T) {
	t.Parallel()

	for name, c := range map[string]struct {
		body string
		user int64
	}{
		"user 0 against a lord 0":    {`[{"id":1,"regionLord":{"name":"a","id":0}}]`, 0},
		"user 0 against a real lord": {`[{"id":1,"regionLord":{"name":"a","id":7}}]`, 0},
		"a negative user":            {`[{"id":1,"regionLord":{"name":"a","id":7}}]`, -7},
		"a lord 0":                   {`[{"id":1,"regionLord":{"name":"a","id":0}},{"id":2,"regionLord":{"name":"b","id":7}}]`, 7},
		"a negative lord":            {`[{"id":1,"regionLord":{"name":"a","id":-7}},{"id":2,"regionLord":{"name":"b","id":7}}]`, 7},
	} {
		if held, err := RegionLord(t.Context(), regionsOnce(c.body), c.user); err == nil {
			t.Errorf("%s: resolved held=%v without error, want the id refused", name, held)
		}
	}
}

// TestARegionResponseWithNoLordAnywhereIsReported is #234 N2: a non-empty
// response in which no region carries a lord is an error, not "not held".
func TestARegionResponseWithNoLordAnywhereIsReported(t *testing.T) {
	t.Parallel()

	held, err := RegionLord(t.Context(), regionsOnce(`[{"id":1},{"id":2,"regionLord":null}]`), 7)
	if err == nil {
		t.Errorf("a response naming no lord resolved held=%v with no error, want it reported", held)
	}
}
