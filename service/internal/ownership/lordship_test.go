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
